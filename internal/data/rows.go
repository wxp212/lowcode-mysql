package data

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "lowcode-mysql/api/lowcode/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gogap/dbstruct"
	"google.golang.org/protobuf/types/known/structpb"
)

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

func (r *RowsRepo) RowsGet(ctx context.Context, req *v1.RowsGetRequest) (*v1.RowsGetReply, error) {
	var dbTable dbstruct.DbTable
	var err error
	if req.Columns == "" {
		dbTable, err = r.data.ds.Describe(req.Table)
		if err != nil {
			return nil, err
		}
	} else {
		query := fmt.Sprintf("select %s from %s", req.Columns, req.Table)
		dbTable, err = r.data.ds.DescribeQuery(query)
		if err != nil {
			return nil, err
		}
	}

	sts, err := dbTable.NewStructSlice()
	if err != nil {
		return nil, err
	}

	r.data.gormdb.Table(req.Table).Find(&sts)
	r.log.Info("RowsGet:", sts)

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

	var reply v1.RowsGetReply
	for _, stmap := range sts_map {
		pbs, err := structpb.NewStruct(stmap)
		if err != nil {
			return nil, err
		}
		reply.Rows = append(reply.Rows, pbs)
	}

	return &reply, nil
}
