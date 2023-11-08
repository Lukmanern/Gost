package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/response"

	fileService "github.com/Lukmanern/gost/service/file"
)

type DevController interface {
	// Ping-ing database 5 times
	PingDatabase(c *fiber.Ctx) error

	// Ping-ing redis 5 times
	PingRedis(c *fiber.Ctx) error

	// Developing Panic handler with defer func()
	Panic(c *fiber.Ctx) error

	// Storing data{key:value} to redis
	StoringToRedis(c *fiber.Ctx) error

	// Getting data from redis
	GetFromRedis(c *fiber.Ctx) error

	// Checking middleware for new role
	CheckNewRole(c *fiber.Ctx) error

	// Checking middleware for new role
	CheckNewPermission(c *fiber.Ctx) error

	// Uploading file into Supabase Bucket
	// See : https://supabase.com/docs/guides/storage
	UploadFile(c *fiber.Ctx) error

	// Removing file from Supabase Bucket
	// See : https://supabase.com/docs/guides/storage
	RemoveFile(c *fiber.Ctx) error

	// Get list file/s from Supabase Bucket
	// See : https://supabase.com/docs/guides/storage
	GetFilesList(c *fiber.Ctx) error
}

type DevControllerImpl struct {
	fileSvc fileService.FileService
	redis   *redis.Client
	db      *gorm.DB
}

var (
	devImpl     *DevControllerImpl
	devImplOnce sync.Once
)

func NewDevControllerImpl() DevController {
	devImplOnce.Do(func() {
		devImpl = &DevControllerImpl{
			fileSvc: fileService.NewFileService(),
			redis:   connector.LoadRedisDatabase(),
			db:      connector.LoadDatabase(),
		}
	})

	return devImpl
}

func (ctr *DevControllerImpl) PingDatabase(c *fiber.Ctx) error {
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

func (ctr *DevControllerImpl) PingRedis(c *fiber.Ctx) error {
	redis := ctr.redis
	if redis == nil {
		return response.Error(c, constants.RedisNil)
	}
	for i := 0; i < 5; i++ {
		status := redis.Ping()
		if status.Err() != nil {
			return response.Error(c, "failed to ping-redis")
		}
	}

	return response.CreateResponse(c, fiber.StatusOK, true, "success ping-redis", nil)
}

func (ctr *DevControllerImpl) Panic(c *fiber.Ctx) error {
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

func (ctr *DevControllerImpl) StoringToRedis(c *fiber.Ctx) error {
	redis := ctr.redis
	if redis == nil {
		return response.Error(c, constants.RedisNil)
	}
	redisStatus := redis.Set("example-key", "example-value", 50*time.Minute)
	if redisStatus.Err() != nil {
		message := fmt.Sprintf("redis status error (%s)", redisStatus.Err().Error())
		return response.Error(c, message)
	}

	return response.SuccessCreated(c, nil)
}

func (ctr *DevControllerImpl) GetFromRedis(c *fiber.Ctx) error {
	redis := ctr.redis
	if redis == nil {
		return response.Error(c, constants.RedisNil)
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

func (ctr *DevControllerImpl) CheckNewRole(c *fiber.Ctx) error {
	return response.CreateResponse(c, fiber.StatusOK, true, "success check new role", nil)
}

func (ctr *DevControllerImpl) CheckNewPermission(c *fiber.Ctx) error {
	return response.CreateResponse(c, fiber.StatusOK, true, "success check new permission", nil)
}

func (ctr *DevControllerImpl) UploadFile(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "failed to parse form file: "+err.Error())
	}
	if file == nil {
		return response.BadRequest(c, "file is nil or not found")
	}
	mimeType := file.Header.Get(fiber.HeaderContentType)
	if mimeType != "application/pdf" {
		return response.BadRequest(c, "only PDF file are allowed for upload")
	}
	maxSize := int64(3 * 1024 * 1024) // 3MB in bytes
	if file.Size > maxSize {
		return response.BadRequest(c, "file size exceeds the maximum allowed (3MB)")
	}

	fileUrl, uploadErr := ctr.fileSvc.UploadFile(file)
	if uploadErr != nil {
		fiberErr, ok := uploadErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+uploadErr.Error())
	}
	return response.SuccessCreated(c, map[string]any{
		"file_url": fileUrl,
	})
}

func (ctr *DevControllerImpl) RemoveFile(c *fiber.Ctx) error {
	var fileName struct {
		FileName string `validate:"required,min=4,max=150" json:"file_name"`
	}
	if err := c.BodyParser(&fileName); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}
	validate := validator.New()
	if err := validate.Struct(&fileName); err != nil {
		return response.BadRequest(c, "invalid json body: "+err.Error())
	}

	removeErr := ctr.fileSvc.RemoveFile(fileName.FileName)
	if removeErr != nil {
		fiberErr, ok := removeErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+removeErr.Error())
	}
	return response.SuccessNoContent(c)
}

func (ctr *DevControllerImpl) GetFilesList(c *fiber.Ctx) error {
	resp, getErr := ctr.fileSvc.GetFilesList()
	if getErr != nil {
		fiberErr, ok := getErr.(*fiber.Error)
		if ok {
			return response.CreateResponse(c, fiberErr.Code, false, fiberErr.Message, nil)
		}
		return response.Error(c, constants.ServerErr+getErr.Error())
	}
	return response.SuccessLoaded(c, resp)
}
