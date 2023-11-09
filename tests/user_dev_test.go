// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package test

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Lukmanern/gost/application"
	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user_management"
	service "github.com/Lukmanern/gost/service/user_management"
)

var (
	userDevService    service.UserManagementService
	userDevController controller.UserManagementController
	appUrl            string
)

func init() {
	env.ReadConfig("./../.env")
	config := env.Configuration()
	appUrl = config.AppUrl
	dbURI := config.GetDatabaseURI()
	privKey := config.GetPrivateKey()
	pubKey := config.GetPublicKey()
	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	connector.LoadDatabase()
	r := connector.LoadRedisCache()
	r.FlushAll() // clear all key:value in redis

	userDevService = service.NewUserManagementService()
	userDevController = controller.NewUserManagementController(userDevService)
}

func TestCreate(t *testing.T) {
	go application.RunApp()
	time.Sleep(4 * time.Second)

	ctr := userDevController
	if ctr == nil {
		t.Error("should not nil")
	}
	c := helper.NewFiberCtx()
	if ctr == nil || c == nil {
		t.Error(constants.ShouldNotErr)
	}

	ctx := c.Context()
	if ctx == nil {
		t.Error("should not nil")
	}
	modelUserCreate := model.UserCreate{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(11),
		IsAdmin:  true,
	}
	userID, createErr := userDevService.Create(ctx, modelUserCreate)
	if createErr != nil || userID < 1 {
		t.Error("should not error and got id more or equal than 1")
	}
	defer func() {
		userDevService.Delete(ctx, userID)
	}()

	testCases := []struct {
		CaseName     string
		payload      model.UserCreate
		resp         response.Response
		wantHttpCode int
	}{
		{
			CaseName: "failed create: email has already used",
			payload: model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    modelUserCreate.Email,
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			resp: response.Response{
				Data:    nil,
				Success: false,
				Message: "",
			},
			wantHttpCode: http.StatusBadRequest,
		},
		{
			CaseName: "success create",
			payload: model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			resp: response.Response{
				Data:    nil,
				Success: true,
				Message: response.MessageSuccessCreated,
			},
			wantHttpCode: http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		req, httpReqErr := http.NewRequest(http.MethodPost, appUrl+"user-management/create", strings.NewReader(string(jsonObject)))
		if httpReqErr != nil {
			t.Fatal("should not nil")
		}
		req.Close = true
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		client := &http.Client{
			Transport: &http.Transport{},
		}
		resp, clientErr := client.Do(req)
		if clientErr != nil {
			t.Fatalf("HTTP request failed: %v", clientErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.wantHttpCode {
			t.Error(constants.ShouldEqual, "but got", resp.StatusCode)
		}

		respModel := response.Response{}
		decodeErr := json.NewDecoder(resp.Body).Decode(&respModel)
		if decodeErr != nil {
			t.Error(constants.ShouldNotErr, decodeErr)
		}

		if tc.resp.Success != respModel.Success {
			t.Error(constants.ShouldEqual, "but got", resp.StatusCode)
		}
		if tc.resp.Message != "" {
			if tc.resp.Message != respModel.Message {
				t.Error(constants.ShouldEqual)
			}
		}
		if tc.resp.Data != nil {
			if !reflect.DeepEqual(tc.resp.Data, respModel.Data) {
				t.Error(constants.ShouldEqual)
			}
		}
		if respModel.Success {
			userByEmail, getErr := userDevService.GetByEmail(ctx, tc.payload.Email)
			if userByEmail.Name != helper.ToTitle(tc.payload.Name) {
				t.Error(constants.ShouldEqual)
			}
			if getErr != nil {
				t.Error(constants.ShouldNotErr, getErr)
			}
			deleteErr := userDevService.Delete(ctx, userByEmail.ID)
			if deleteErr != nil {
				t.Error(constants.ShouldNotErr, deleteErr)
			}
		}
	}
}

// Create(c *fiber.Ctx) error
// Get(c *fiber.Ctx) error
// GetAll(c *fiber.Ctx) error
// Update(c *fiber.Ctx) error
// Delete(c *fiber.Ctx) error
