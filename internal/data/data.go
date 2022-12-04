package data

import (
	"lowcode-mysql/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/gogap/dbstruct"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewRowsRepo)

// Data .
type Data struct {
	ds     *dbstruct.DBStruct
	gormdb *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	ds, err := dbstruct.New(
		dbstruct.DataSource(c.Database.Driver, c.Database.Source),
		dbstruct.CreateTabelDSN(c.Database.Source),
	)
	if err != nil {
		log.NewHelper(logger).Fatal(err)
	}

	gormdb, err := gorm.Open(mysql.Open(ds.Options.DSN), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		log.NewHelper(logger).Fatal(err)
	}

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		ds:     ds,
		gormdb: gormdb,
	}, cleanup, nil
}
