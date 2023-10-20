package controller

import (
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
)

type DevController interface {
	BitfieldTesting(c *fiber.Ctx) error
	PingDatabase(c *fiber.Ctx) error
	PingRedis(c *fiber.Ctx) error
	Panic(c *fiber.Ctx) error
	NewJWT(c *fiber.Ctx) error
}

type DevControllerImpl struct{}

var (
	devImpl     *DevControllerImpl
	devImplOnce sync.Once
)

func NewDevControllerImpl() DevController {
	devImplOnce.Do(func() {
		devImpl = &DevControllerImpl{}
	})

	return devImpl
}

// DevControllerelopement Process
func (ctr DevControllerImpl) BitfieldTesting(c *fiber.Ctx) error {
	return response.CreateResponse(c, fiber.StatusOK, true, "Success Bitfield DevController", nil)
}

func (ctr DevControllerImpl) PingDatabase(c *fiber.Ctx) error {
	db := connector.LoadDatabase()
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
	rds := connector.LoadRedisDatabase()
	for i := 0; i < 5; i++ {
		status := rds.Ping()
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
			return response.Error(c, "message panic: "+r.(string))
		}
		return nil
	}()
	panic("Panic message") // message should string
}

func (ctr DevControllerImpl) NewJWT(c *fiber.Ctx) error {
	defer func() error {
		r := recover()
		if r != nil {
			return response.Error(c, "message panic: "+r.(string))
		}
		return nil
	}()
	panic("Testing New JWT") // message should string
}
