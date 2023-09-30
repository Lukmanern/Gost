package connection

import (
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	gormDatabase     *gorm.DB
	gormDatabaseOnce sync.Once
)

func LoadDatabase() *gorm.DB {
	gormDatabaseOnce.Do(func() {
		// try to connect to database
		dsn := "root:@tcp(localhost:3306)/gost?charset=utf8mb4&multiStatements=true&parseTime=true"
		var conErr error

		gormDatabase, conErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if conErr != nil {
			log.Panicf("can't connect to postgres %s", conErr)
		}
		if gormDatabase == nil {
			log.Panic("database is null")
		}
		gormDatabase.Logger = logger.Default.LogMode(logger.Info)

		// try to ping
		sqlDB, sqlErr := gormDatabase.DB()
		if sqlErr != nil {
			log.Panicf("can't get sql-db %s", sqlErr)
		}
		pingErr := sqlDB.Ping()
		if pingErr != nil {
			log.Panicf("can't ping sql-db %s", pingErr)
		}
	})

	return gormDatabase
}
