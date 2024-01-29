package controller

import (
	"log"
	"testing"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/role"
	service "github.com/Lukmanern/gost/service/role"
	"github.com/stretchr/testify/assert"
)

const (
	headerTestName string = "at Role Controller Test"
)

var (
	baseURL  string
	token    string
	timeNow  time.Time
	roleRepo repository.RoleRepository
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	config := env.Configuration()
	baseURL = config.AppURL
	token = helper.GenerateToken()
	timeNow = time.Now()
	roleRepo = repository.NewRoleRepository()

	connector.LoadDatabase()
	r := connector.LoadRedisCache()
	r.FlushAll() // clear all key:value in redis
}

func TestCreate(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	_ = createRole()

	// pathURL := "role"
	// URL := baseURL + pathURL
	// for _, tc := range testCases {
	// 	log.Println(tc.Name, headerTestName)

	// 	// Marshal payload to JSON
	// 	jsonData, marshalErr := json.Marshal(&tc.Payload)
	// 	assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

	// 	// Create HTTP request
	// 	req := httptest.NewRequest(fiber.MethodPost, URL, bytes.NewReader(jsonData))
	// 	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// 	// Set up Fiber app and handle the request with the controller
	// 	app := fiber.New()
	// 	app.Post(pathURL, controller.Create)
	// 	req.Close = true

	// 	// run test
	// 	res, testErr := app.Test(req, -1)
	// 	assert.Nil(t, testErr, consts.ShouldNil, testErr)
	// 	defer res.Body.Close()
	// 	assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, res.StatusCode)

	// 	if res.StatusCode == fiber.StatusCreated {
	// 		payload, ok := tc.Payload.(model.RoleCreate)
	// 		assert.True(t, ok, "should true", headerTestName)
	// 		log.Println(payload)
	// 		entityRole, getErr := repository.GetByID(ctx, payload.Email)
	// 		assert.NoError(t, getErr, consts.ShouldNotErr, headerTestName)
	// 		deleteErr := repository.Delete(ctx, entityRole.ID)
	// 		assert.NoError(t, deleteErr, consts.ShouldNotErr, headerTestName)
	// 	}

	// 	if res.StatusCode != fiber.StatusNoContent {
	// 		responseStruct := response.Response{}
	// 		err := json.NewDecoder(res.Body).Decode(&responseStruct)
	// 		assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
	// 	}
	// }
}

func TestGet(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	_ = createRole()
}

func TestGetAll(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	_ = createRole()
}

func TestUpdate(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	_ = createRole()
}

func TestDelete(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	_ = createRole()
}

func createRole() entity.Role {
	repo := roleRepo
	ctx := helper.NewFiberCtx().Context()
	data := entity.Role{
		Name:        helper.RandomString(15),
		Description: helper.RandomWords(10),
	}
	data.SetCreateTime()
	id, err := repo.Create(ctx, data)
	if err != nil {
		log.Fatal("Failed create user", headerTestName)
	}
	data.ID = id
	return data
}
