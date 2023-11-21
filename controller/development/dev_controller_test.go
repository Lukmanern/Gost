package controller

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

const (
	TestName    = "Development Controller Test"
	FilePath    = "./controller/development"
	AddTestName = ", at " + TestName + " in " + FilePath
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestNewDevControllerImpl(t *testing.T) {
	assert := assert.New(t)
	ctr := NewDevControllerImpl()
	c := helper.NewFiberCtx()
	assert.NotNil(ctr, "Expected NewDevControllerImpl to not be nil"+AddTestName)
	assert.NotNil(c, "Expected FiberCtx to not be nil"+AddTestName)
}

type TestCase struct {
	Name             string
	Handler          func(*fiber.Ctx) error
	AdditionalAction func()
	Payload          map[string]any
	ResponseCode     int
	ResponseBody     response.Response
}

func TestAllControllers(t *testing.T) {
	assert := assert.New(t)
	ctr := NewDevControllerImpl()
	c := helper.NewFiberCtx()
	assert.NotNil(ctr, "Expected NewDevControllerImpl to not be nil"+AddTestName)
	assert.NotNil(c, "Expected FiberCtx to not be nil"+AddTestName)

	testCases := []TestCase{
		{
			Name:             "PingDB" + AddTestName,
			Handler:          ctr.PingDatabase,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:             "PingRedis" + AddTestName,
			Handler:          ctr.PingRedis,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:             "Panic" + AddTestName,
			Handler:          ctr.Panic,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusInternalServerError,
			ResponseBody: response.Response{
				Success: false,
			},
		},
		{
			Name:    "GetFromRedis-1" + AddTestName,
			Handler: ctr.GetFromRedis,
			AdditionalAction: func() {
				connector.LoadRedisCache().FlushAll()
			},
			ResponseCode: fiber.StatusInternalServerError,
			ResponseBody: response.Response{
				Success: false,
			},
		},
		{
			Name:             "StoringToRedis" + AddTestName,
			Handler:          ctr.StoringToRedis,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusCreated,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:             "GetFromRedis-2" + AddTestName,
			Handler:          ctr.GetFromRedis,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:    "GetFromRedis-3" + AddTestName,
			Handler: ctr.GetFromRedis,
			AdditionalAction: func() {
				connector.LoadRedisCache().FlushAll()
			},
			ResponseCode: fiber.StatusInternalServerError,
			ResponseBody: response.Response{
				Success: false,
			},
		},
		{
			Name:             "CheckNewRole" + AddTestName,
			Handler:          ctr.CheckNewRole,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:             "CheckNewPermission" + AddTestName,
			Handler:          ctr.CheckNewPermission,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:             "UploadFile" + AddTestName,
			Handler:          ctr.UploadFile,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusBadRequest,
			ResponseBody: response.Response{
				Success: false,
			},
		},
		{
			Name:             "RemoveFile" + AddTestName,
			Handler:          ctr.RemoveFile,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusBadRequest,
			ResponseBody: response.Response{
				Success: false,
			},
		},
		{
			Name:             "GetFilesList" + AddTestName,
			Handler:          ctr.GetFilesList,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
		{
			Name:             "FakeHandler" + AddTestName,
			Handler:          ctr.FakeHandler,
			AdditionalAction: nil,
			ResponseCode:     fiber.StatusOK,
			ResponseBody: response.Response{
				Success: true,
			},
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name)
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		c.Request().Header.SetMethod(fiber.MethodGet)

		if tc.AdditionalAction != nil {
			tc.AdditionalAction()
		}

		tc.Handler(c)
		res := c.Response()
		assert.Equal(res.StatusCode(), tc.ResponseCode, constants.ShouldEqual, res.StatusCode())

		resBody := res.Body()
		resString := string(resBody)
		resBodyStruct := response.Response{}
		err := json.Unmarshal([]byte(resString), &resBodyStruct)
		assert.Nilf(err, "failed to parse response JSON: %v", err)
		assert.Equal(tc.ResponseBody.Success, resBodyStruct.Success, "expected success value should match")
	}
}
