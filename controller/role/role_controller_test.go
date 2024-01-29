package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	repository "github.com/Lukmanern/gost/repository/role"
	service "github.com/Lukmanern/gost/service/role"
	"github.com/gofiber/fiber/v2"
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

func TestUnauthorized(t *testing.T) {
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)

	handlers := []func(c *fiber.Ctx) error{
		controller.Get,
		controller.GetAll,
		controller.Create,
		controller.Update,
		controller.Delete,
	}
	for _, handler := range handlers {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		handler(c)
		res := c.Response()
		assert.Equalf(t, res.StatusCode(), fiber.StatusUnauthorized, "Expected response code %d, but got %d", fiber.StatusUnauthorized, res.StatusCode())
	}
}

func TestCreate(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validRole := createRole()
	defer repository.Delete(ctx, validRole.ID)

	type testCase struct {
		Name    string
		ResCode int
		Payload model.RoleCreate
	}

	testCases := []testCase{
		{
			Name:    "Success Create Role -1",
			ResCode: fiber.StatusCreated,
			Payload: model.RoleCreate{
				Name:        strings.ToLower(helper.RandomString(14)),
				Description: helper.RandomWords(10),
			},
		},
		{
			Name:    "Success Create Role -2",
			ResCode: fiber.StatusCreated,
			Payload: model.RoleCreate{
				Name:        strings.ToLower(helper.RandomString(14)),
				Description: helper.RandomWords(10),
			},
		},
		{
			Name:    "Failed Create Role -1: invalid name / name is already used",
			ResCode: fiber.StatusBadRequest,
			Payload: model.RoleCreate{
				Name:        validRole.Name,
				Description: helper.RandomWords(10),
			},
		},
		{
			Name:    "Failed Create Role -2: invalid name / name too short",
			ResCode: fiber.StatusBadRequest,
			Payload: model.RoleCreate{
				Name:        "",
				Description: helper.RandomWords(10),
			},
		},
		{
			Name:    "Failed Create Role -3: invalid name / name too long",
			ResCode: fiber.StatusBadRequest,
			Payload: model.RoleCreate{
				Name:        helper.RandomWords(100),
				Description: helper.RandomWords(10),
			},
		},
	}

	pathURL := "role"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Marshal payload to JSON
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPost, URL, bytes.NewReader(jsonData))
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Post(pathURL, jwtHandler.IsAuthenticated, controller.Create)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}

			if res.StatusCode == fiber.StatusCreated {
				data, ok1 := data.Data.(map[string]interface{})
				if !ok1 {
					t.Error("should ok1")
				}
				anyID, ok2 := data["id"]
				if !ok2 {
					t.Error("should ok2")
				}
				intID, ok3 := anyID.(float64)
				if !ok3 {
					t.Error("should ok3")
				}
				deleteErr := repository.Delete(ctx, int(intID))
				if deleteErr != nil {
					t.Error(consts.ShouldNotErr)
				}
			}
		}
	}
}

func TestGet(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validRole := createRole()
	defer repository.Delete(ctx, validRole.ID)

	type testCase struct {
		Name    string
		ResCode int
		ID      string
	}

	testCases := []testCase{
		{
			Name:    "Success Create Role -1",
			ResCode: fiber.StatusOK,
			ID:      strconv.Itoa(validRole.ID),
		},
		{
			Name:    "Failed Create Role -1: invalid ID",
			ResCode: fiber.StatusBadRequest,
			ID:      strconv.Itoa(-1),
		},
		{
			Name:    "Failed Create Role -2: invalid ID",
			ResCode: fiber.StatusBadRequest,
			ID:      "invalid-id",
		},
		{
			Name:    "Failed Create Role -3: not found",
			ResCode: fiber.StatusNotFound,
			ID:      strconv.Itoa(validRole.ID + 99),
		},
	}

	for _, tc := range testCases {
		pathURL := "role/" // "/role/:id"
		URL := baseURL + pathURL + tc.ID
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodGet, URL, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Get(pathURL+":id", jwtHandler.IsAuthenticated, controller.Get)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
	}
}

func TestGetAll(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validRole := createRole()
	defer repository.Delete(ctx, validRole.ID)

	type testCase struct {
		Name    string
		ResCode int
		Params  string
	}

	testCases := []testCase{
		{
			Name:    "Success Get All Role -1",
			ResCode: fiber.StatusOK,
			Params:  "?limit=100&page=1",
		},
		{
			Name:    "Failed Get All Role -1: invalid parameter",
			ResCode: fiber.StatusBadRequest,
			Params:  "?limit=-1&page=1",
		},
		{
			Name:    "Failed Get All Role -2: invalid parameter",
			ResCode: fiber.StatusBadRequest,
			Params:  "?limit=100&page=-10",
		},
	}

	for _, tc := range testCases {
		pathURL := "role/"
		URL := baseURL + pathURL + tc.Params
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodGet, URL, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Get(pathURL, jwtHandler.IsAuthenticated, controller.GetAll)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
	}
}

func TestUpdate(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validRole := createRole()
	defer repository.Delete(ctx, validRole.ID)

	type testCase struct {
		Name    string
		ResCode int
		ID      string
		Payload model.RoleUpdate
	}

	testCases := []testCase{
		{
			Name:    "Success Update Role -1",
			ResCode: fiber.StatusNoContent,
			ID:      strconv.Itoa(validRole.ID),
			Payload: model.RoleUpdate{
				Name:        strings.ToLower(helper.RandomString(10)),
				Description: helper.RandomWords(10),
			},
		},
		{
			Name:    "Success Update Role -2",
			ResCode: fiber.StatusNoContent,
			ID:      strconv.Itoa(validRole.ID),
			Payload: model.RoleUpdate{
				Name:        strings.ToLower(helper.RandomString(20)),
				Description: helper.RandomWords(5),
			},
		},
		{
			Name:    "Failed Update Role -1: name too long",
			ResCode: fiber.StatusBadRequest,
			ID:      strconv.Itoa(validRole.ID),
			Payload: model.RoleUpdate{
				Name:        strings.ToLower(helper.RandomWords(50)),
				Description: helper.RandomWords(5),
			},
		},
		{
			Name:    "Failed Update Role -2: name too short",
			ResCode: fiber.StatusBadRequest,
			ID:      strconv.Itoa(validRole.ID),
			Payload: model.RoleUpdate{
				Name:        "",
				Description: helper.RandomWords(5),
			},
		},
		{
			Name:    "Failed Update Role -2: data not found",
			ResCode: fiber.StatusNotFound,
			ID:      strconv.Itoa(validRole.ID + 999),
			Payload: model.RoleUpdate{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
		},
		{
			Name:    "Failed Update Role -2: invalid ID",
			ResCode: fiber.StatusBadRequest,
			ID:      strconv.Itoa(-10),
			Payload: model.RoleUpdate{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
		},
	}

	for _, tc := range testCases {
		pathURL := "role/"
		URL := baseURL + pathURL + tc.ID
		log.Println(tc.Name, headerTestName)

		// Marshal payload to JSON
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPut, URL, bytes.NewReader(jsonData))
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Put(pathURL+":id", jwtHandler.IsAuthenticated, controller.Update)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
		if res.StatusCode == fiber.StatusNoContent {
			id, _ := strconv.Atoi(tc.ID)
			enttRole, getErr := repository.GetByID(ctx, id)
			assert.Nil(t, getErr, consts.ShouldNil, testErr, tc.Name, headerTestName)
			assert.Equal(t, enttRole.Name, tc.Payload.Name, consts.ShouldEqual, tc.Name, headerTestName)
			assert.Equal(t, enttRole.Description, tc.Payload.Description, consts.ShouldEqual, tc.Name, headerTestName)
		}
	}
}

func TestDelete(t *testing.T) {
	repository := repository.NewRoleRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewRoleController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validRole := createRole()
	defer repository.Delete(ctx, validRole.ID)

	type testCase struct {
		Name    string
		ResCode int
		ID      string
	}

	testCases := []testCase{
		{
			Name:    "Success Delete Role -1",
			ResCode: fiber.StatusNoContent,
			ID:      strconv.Itoa(validRole.ID),
		},
		{
			Name:    "Failed Delete Role -1: data not found / already deleted",
			ResCode: fiber.StatusNotFound,
			ID:      strconv.Itoa(validRole.ID),
		},
		{
			Name:    "Failed Delete Role -2: data not found",
			ResCode: fiber.StatusNotFound,
			ID:      strconv.Itoa(validRole.ID + 999),
		},
		{
			Name:    "Failed Delete Role -3: invalid ID",
			ResCode: fiber.StatusBadRequest,
			ID:      "invalid-id",
		},
		{
			Name:    "Failed Delete Role -4: invalid ID",
			ResCode: fiber.StatusBadRequest,
			ID:      "-10",
		},
	}

	for _, tc := range testCases {
		pathURL := "role/"
		URL := baseURL + pathURL + tc.ID
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodDelete, URL, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Delete(pathURL+":id", jwtHandler.IsAuthenticated, controller.Delete)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
	}
}

func createRole() entity.Role {
	repo := roleRepo
	ctx := helper.NewFiberCtx().Context()
	data := entity.Role{
		Name:        strings.ToLower(helper.RandomString(15)),
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
