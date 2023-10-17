package connector

import (
	"log"
	"sync"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	gormDatabase     *gorm.DB
	gormDatabaseOnce sync.Once

	redisDatastore     *redis.Client
	redisDatastoreOnce sync.Once
)

// MySQL
func LoadDatabase() *gorm.DB {
	gormDatabaseOnce.Do(func() {
		// try to connect to database
		env.ReadConfig("./.env")
		config := env.Configuration()
		dsn := config.GetDatabaseURI()
		var conErr error

		gormDatabase, conErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if conErr != nil {
			log.Panicf("can't connect to database %s", conErr)
		}
		if gormDatabase == nil {
			log.Panic("database is null")
		}
		gormDatabase.Logger = logger.Default.LogMode(logger.Info)

		// try to ping the database
		database, sqlErr := gormDatabase.DB()
		if sqlErr != nil {
			log.Panicf("can't get sql-db : %s", sqlErr)
		}
		pingErr := database.Ping()
		if pingErr != nil {
			log.Panicf("can't ping sql-db : %s", pingErr)
		}
	})

	return gormDatabase
}

// Redis
func LoadRedisDatabase() *redis.Client {
	redisDatastoreOnce.Do(func() {
		env.ReadConfig("./.env")
		config := env.Configuration()
		opt, err := redis.ParseURL(config.RedisURI)
		if err != nil {
			log.Panicf("cannot connect to redis %s", err)
		}

		redisDatastore = redis.NewClient(opt)

		_, err = redisDatastore.Ping().Result()
		if err != nil {
			log.Panicf("cannot ping to redis %T: %s", config.RedisURI, err)
		}
	})

	return redisDatastore
}
