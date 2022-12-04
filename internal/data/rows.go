package data

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	v1 "lowcode-mysql/api/lowcode/v1"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gogap/dbstruct"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type ForeignKey struct {
	TableName            string `gorm:"column:TABLE_NAME"`
	ColumnName           string `gorm:"column:COLUMN_NAME"`
	ReferencedTableName  string `gorm:"column:REFERENCED_TABLE_NAME"`
	ReferencedColumnName string `gorm:"column:REFERENCED_COLUMN_NAME"`
}

type Layer struct {
	Table   string
	Columns string // table.column1,table.column2
	Next    map[string]*Layer
}

type Conditions struct {
	DataBase string
	Tables   string // table1,table2
	Where    string
}

var layers map[string]*Layer = make(map[string]*Layer)

type RowsRepo struct {
	data *Data
	log  *log.Helper
}

func NewRowsRepo(data *Data, logger log.Logger) *RowsRepo {
	return &RowsRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *RowsRepo) RowsCreate(ctx context.Context, req *v1.RowsCreateRequest) (*v1.RowsCreateReply, error) {
	dbTable, err := r.data.ds.Describe(req.Table)
	if err != nil {
		return nil, err
	}

	reply := &v1.RowsCreateReply{}

	tx := r.data.gormdb.Begin()
	defer tx.Rollback()

	for _, row := range req.Rows {
		st, err := dbTable.NewStruct()
		if err != nil {
			return nil, err
		}

		row_json, err := row.MarshalJSON()
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(row_json, st)
		if err != nil {
			return nil, err
		}

		err = tx.Table(req.Table).Create(st).Error
		if err != nil {
			return nil, err
		}

		result_json, err := json.Marshal(st)
		if err != nil {
			return nil, err
		}
		result := &structpb.Struct{}
		err = protojson.Unmarshal(result_json, result)
		if err != nil {
			return nil, err
		}
		reply.Rows = append(reply.Rows, result)
	}

	tx.Commit()
	return reply, nil
}

/*
 * TODO: select Columns
 * TODO: 防撞库?
 * need to add TableName to lastLayerValue
 */
func (r *RowsRepo) RowsGetLayer(ctx context.Context, curLayer *Layer, lastLayerValue map[string]interface{}, conditions *Conditions) ([]map[string]interface{}, error) {
	var dbTable dbstruct.DbTable
	var err error
	var columns string
	if curLayer.Columns == "" {
		dbTable, err = r.data.ds.Describe(curLayer.Table)
		if err != nil {
			return nil, err
		}
		columns = fmt.Sprintf("%s.*", curLayer.Table)
	} else {
		query := fmt.Sprintf("select %s from %s", curLayer.Columns, curLayer.Table)
		dbTable, err = r.data.ds.DescribeQuery(query)
		if err != nil {
			return nil, err
		}

		columns = curLayer.Columns
	}

	sts, err := dbTable.NewStructSlice()
	if err != nil {
		return nil, err
	}

	var fkWhere string
	if lastLayerValue != nil {
		lastLayerTN, ok := lastLayerValue["TableName"]
		if !ok {
			err_string := "TableName must be added to lastLayerValue"
			r.log.Info(err_string)
			return nil, errors.New(err_string)
		}
		lastLayerTNString := lastLayerTN.(string)
		var foreignKeys []ForeignKey

		r.data.gormdb.Raw(
			`SELECT
			TABLE_NAME,
			COLUMN_NAME,
			REFERENCED_TABLE_NAME,
			REFERENCED_COLUMN_NAME
		FROM
			INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE
			REFERENCED_TABLE_SCHEMA = ?
			AND ((REFERENCED_TABLE_NAME = ? AND TABLE_NAME = ?)
			OR (REFERENCED_TABLE_NAME = ? AND TABLE_NAME = ?))`,
			conditions.DataBase,
			curLayer.Table, lastLayerTNString,
			lastLayerTNString, curLayer.Table,
		).Find(&foreignKeys)

		r.log.Info("RowsGetLayer, foreignKeys:", foreignKeys)
		for _, fk := range foreignKeys {
			if fk.TableName == curLayer.Table {
				rcnNameMapped := r.data.ds.Options.NameMap(fk.ReferencedColumnName)
				id, ok := lastLayerValue[rcnNameMapped]
				if !ok {
					err_string := "RowsGetLayer, Id must be added to lastLayerValue"
					r.log.Info(err_string)
					return nil, errors.New(err_string)
				}
				fkWhere = fmt.Sprintf("%s.%s = %d", fk.TableName, fk.ColumnName, cast.ToUint64(id))
			} else if fk.ReferencedTableName == curLayer.Table {
				fkcnNameMapped := r.data.ds.Options.NameMap(fk.ColumnName)
				fkValue, ok := lastLayerValue[fkcnNameMapped]
				if !ok {
					err_string := "RowsGetLayer, fk value must be added to lastLayerValue"
					r.log.Info(err_string)
					return nil, errors.New(err_string)
				}
				fkWhere = fmt.Sprintf("%s.%s = %d", fk.ReferencedTableName, fk.ReferencedColumnName, cast.ToUint64(fkValue))
			}
		}
	}

	raw := fmt.Sprintf("select distinct %s from %s", columns, conditions.Tables)
	if fkWhere != "" || conditions.Where != "" {
		raw += " where"
	}
	if fkWhere != "" {
		raw += fmt.Sprintf(" (%s)", fkWhere)
	}
	if conditions.Where != "" {
		if fkWhere != "" {
			raw += " and"
		}
		raw += fmt.Sprintf(" (%s)", conditions.Where)
	}
	r.data.gormdb.Raw(raw).Find(&sts)
	r.log.Info("RowsGetLayer:", sts)

	sts_json, err := json.Marshal(sts)
	if err != nil {
		return nil, err
	}
	//r.log.Info("sts_json:", string(sts_json))
	var sts_map []map[string]interface{}
	err = json.Unmarshal(sts_json, &sts_map)
	if err != nil {
		return nil, err
	}

	if curLayer.Next != nil {
		for _, stmap := range sts_map {
			for _, layer := range curLayer.Next {
				stmap["TableName"] = curLayer.Table
				layer_sts_map, err := r.RowsGetLayer(ctx, layer, stmap, conditions)
				if err != nil {
					return nil, err
				}
				stmap[layer.Table] = layer_sts_map
				delete(stmap, "TableName")
			}
		}
	}

	return sts_map, nil
}

func (r *RowsRepo) RowsGet(ctx context.Context, req *v1.RowsGetRequest) (*v1.RowsGetReply, error) {
	conditions := &Conditions{
		DataBase: "test",
		Tables:   req.Table,
		Where:    req.Where,
	}
	tables := make(map[string]string)

	// get layers
	cur_layer := &Layer{
		Table: req.Table,
	}
	layers[req.Table] = cur_layer
	if req.Columns != "" {
		columns := strings.Split(req.Columns, ",")
		//r.log.Info("RowsGet, req columns: ", columns)
		for _, col := range columns {
			colLayers := strings.Split(col, ".")
			//r.log.Info("RowsGet, req colLayers: ", colLayers)
			for i, cl := range colLayers {
				if i == len(colLayers)-1 {
					if cur_layer.Columns == "" {
						cur_layer.Columns = fmt.Sprintf("%s.%s", cur_layer.Table, cl)
					} else {
						cur_layer.Columns = cur_layer.Columns + "," + fmt.Sprintf("%s.%s", cur_layer.Table, cl)
					}
				} else {
					tables[cl] = cl
					if cur_layer.Next == nil {
						cur_layer.Next = make(map[string]*Layer)
					}
					_, ok := cur_layer.Next[cl]
					if !ok {
						new_layer := &Layer{
							Table: cl,
						}
						cur_layer.Next[cl] = new_layer
					}
					cur_layer = cur_layer.Next[cl]
				}
			}
		}
	}
	//r.log.Info("RowsGet, parsed layers:", layers)

	for t := range tables {
		conditions.Tables = conditions.Tables + "," + t
	}

	sts_map, err := r.RowsGetLayer(ctx, layers[req.Table], nil, conditions)
	if err != nil {
		return nil, err
	}

	var reply v1.RowsGetReply
	for _, stmap := range sts_map {
		reply_json, err := json.Marshal(stmap)
		if err != nil {
			return nil, err
		}

		s := &structpb.Struct{}
		err = protojson.Unmarshal(reply_json, s)
		if err != nil {
			return nil, err
		}
		reply.Rows = append(reply.Rows, s)
	}
	return &reply, nil
}
