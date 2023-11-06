package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/response"

	uploadService "github.com/Lukmanern/gost/service/upload_file"
)

type DevController interface {
	PingDatabase(c *fiber.Ctx) error
	PingRedis(c *fiber.Ctx) error
	Panic(c *fiber.Ctx) error
	StoringToRedis(c *fiber.Ctx) error
	GetFromRedis(c *fiber.Ctx) error
	CheckNewRole(c *fiber.Ctx) error
	CheckNewPermission(c *fiber.Ctx) error
	UploadFile(c *fiber.Ctx) error
}

type DevControllerImpl struct {
	redis *redis.Client
	db    *gorm.DB
}

var (
	devImpl     *DevControllerImpl
	devImplOnce sync.Once
)

func NewDevControllerImpl() DevController {
	devImplOnce.Do(func() {
		devImpl = &DevControllerImpl{
			redis: connector.LoadRedisDatabase(),
			db:    connector.LoadDatabase(),
		}
	})

	return devImpl
}

func (ctr DevControllerImpl) PingDatabase(c *fiber.Ctx) error {
	db := ctr.db
	if db == nil {
		return response.Error(c, "failed db is nil")
	}
	sqldb, sqlErr := db.DB()
	if sqlErr != nil {
		return response.Error(c, "failed get sql-db")
	}
	for i := 0; i < 5; i++ {
		pingErr := sqldb.Ping()
		if pingErr != nil {
			return response.Error(c, "failed to ping-sql-db")
		}
	}

	return response.CreateResponse(c, fiber.StatusOK, true, "success ping-sql-db", nil)
}

func (ctr DevControllerImpl) PingRedis(c *fiber.Ctx) error {
	redis := ctr.redis
	if redis == nil {
		return response.Error(c, "redis nil value")
	}
	for i := 0; i < 5; i++ {
		status := redis.Ping()
		if status.Err() != nil {
			return response.Error(c, "failed to ping-redis")
		}
	}

	return response.CreateResponse(c, fiber.StatusOK, true, "success ping-redis", nil)
}

func (ctr DevControllerImpl) Panic(c *fiber.Ctx) error {
	defer func() error {
		r := recover()
		if r != nil {
			message := "message panic: " + r.(string)
			return response.Error(c, message)
		}
		return nil
	}()
	panic("Panic message") // message should string
}

func (ctr DevControllerImpl) StoringToRedis(c *fiber.Ctx) error {
	redis := ctr.redis
	if redis == nil {
		return response.Error(c, "redis nil value")
	}
	redisStatus := redis.Set("example-key", "example-value", 50*time.Minute)
	if redisStatus.Err() != nil {
		message := fmt.Sprintf("redis status error (%s)", redisStatus.Err().Error())
		return response.Error(c, message)
	}

	return response.SuccessCreated(c, nil)
}

func (ctr DevControllerImpl) GetFromRedis(c *fiber.Ctx) error {
	redis := ctr.redis
	if redis == nil {
		return response.Error(c, "redis nil value")
	}
	redisStatus := redis.Get("example-key")
	if redisStatus.Err() != nil {
		message := fmt.Sprintf("redis status error (%s)", redisStatus.Err().Error())
		return response.Error(c, message)
	}
	res, resErr := redisStatus.Result()
	if resErr != nil {
		message := fmt.Sprintf("redis result error (%s)", resErr.Error())
		return response.Error(c, message)
	}

	return response.SuccessLoaded(c, res)
}

func (ctr DevControllerImpl) CheckNewRole(c *fiber.Ctx) error {
	return response.CreateResponse(c, fiber.StatusOK, true, "success check new role", nil)
}

func (ctr DevControllerImpl) CheckNewPermission(c *fiber.Ctx) error {
	return response.CreateResponse(c, fiber.StatusOK, true, "success check new permission", nil)
}

func (ctr DevControllerImpl) UploadFile(c *fiber.Ctx) error {
	service := uploadService.NewClient("", "", "")
	_, err := service.Upload(nil)
	if err != nil {
		return response.Error(c, "internal server error: "+err.Error())
	}

	return response.SuccessCreated(c, nil)
}
