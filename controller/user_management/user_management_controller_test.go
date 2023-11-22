package controller_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	controller "github.com/Lukmanern/gost/controller/user_management"
	service "github.com/Lukmanern/gost/service/user_management"
)

const (
	testName    = "User Management Controller Test"
	filePath    = "./controller/user_management"
	addTestName = ", at " + testName + " in " + filePath
)

var (
	userDevService    service.UserManagementService
	userDevController controller.UserManagementController
	appURL            string
)

func init() {
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appURL = config.AppURL

	connector.LoadDatabase()
	connector.LoadRedisCache()

	userDevService = service.NewUserManagementService()
	userDevController = controller.NewUserManagementController(userDevService)
}

func TestCreate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userDevController
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser := model.UserCreate{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(11),
	}
	createdUserID, createErr := userDevService.Create(c.Context(), createdUser)
	if createErr != nil || createdUserID < 1 {
		t.Fatal("should not error and userID should more tha zero")
	}
	defer func() {
		userDevService.Delete(c.Context(), createdUserID)
		r := recover()
		if r != nil {
			t.Error("panic ::", r)
		}
	}()

	testCases := []struct {
		Name    string
		Payload *model.UserCreate
		ResCode int
	}{
		{
			Name: "success create user -1",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusCreated,
		},
		{
			Name: "success create user -2",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusCreated,
		},
		{
			Name: "success create user -3",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusCreated,
		},
		{
			Name: "failed create user: invalid email address",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    "invalid-email-address",
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusBadRequest,
		},
		{
			Name: "failed create user: email already used",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    createdUser.Email,
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusBadRequest,
		},
		{
			Name: "failed create user: password too short",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: "s-;",
				IsAdmin:  true,
			},
			ResCode: fiber.StatusBadRequest,
		},
		{
			Name:    "failed create user: nil Payload, validate failed",
			Payload: nil,
			ResCode: fiber.StatusBadRequest,
		},
	}

	endpoint := "/user-management/"
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endpoint, ctr.Create)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestGetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userDevController
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	userIDs := make([]int, 0)
	for i := 0; i < 10; i++ {
		createdUser := model.UserCreate{
			Name:     helper.RandomString(11),
			Email:    helper.RandomEmail(),
			Password: helper.RandomString(11),
			IsAdmin:  true,
		}
		createdUserID, createErr := userDevService.Create(ctx, createdUser)
		if createErr != nil || createdUserID <= 0 {
			t.Error("should not error and more than zero")
		}
		userIDs = append(userIDs, createdUserID)
	}

	defer func() {
		for _, id := range userIDs {
			userDevService.Delete(ctx, id)
		}
		r := recover()
		if r != nil {
			t.Error("panic ::", r)
		}
	}()

	testCases := []struct {
		Name    string
		Payload string
		ResCode int
	}{
		{
			Name:    "success getall",
			Payload: "page=1&limit=100&search=",
			ResCode: http.StatusOK,
		},
		{
			Name:    "success getall",
			Payload: "page=2&limit=10&search=",
			ResCode: http.StatusOK,
		},
		{
			Name:    "success getall",
			Payload: "page=3&limit=10&search=",
			ResCode: http.StatusOK,
		},
		{
			Name:    "failed getall",
			Payload: "page=-1&limit=-100&search=",
			ResCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/user-management?"+tc.Payload, nil)
		app := fiber.New()
		app.Get("/user-management", userDevController.GetAll)
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr, err.Error())
		}
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestUpdate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userDevController
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser := model.UserCreate{
		Name:     helper.RandomString(11),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(11),
		IsAdmin:  true,
	}
	createdUserID, createErr := userDevService.Create(ctx, createdUser)
	if createErr != nil || createdUserID <= 0 {
		t.Error("should not error and more than zero")
	}
	defer func() {
		userDevService.Delete(ctx, createdUserID)
		r := recover()
		if r != nil {
			t.Error("panic ::", r)
		}
	}()

	testCases := []struct {
		Name    string
		Payload *model.UserProfileUpdate
		ResCode int
	}{
		{
			Name: "success update user -1",
			Payload: &model.UserProfileUpdate{
				ID:   createdUserID,
				Name: helper.RandomString(6),
			},
			ResCode: http.StatusNoContent,
		},
		{
			Name: "success update user -2",
			Payload: &model.UserProfileUpdate{
				ID:   createdUserID,
				Name: helper.RandomString(8),
			},
			ResCode: http.StatusNoContent,
		},
		{
			Name: "success update user -3",
			Payload: &model.UserProfileUpdate{
				ID:   createdUserID,
				Name: helper.RandomString(10),
			},
			ResCode: http.StatusNoContent,
		},
		{
			Name:    "failed update: invalid id",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserProfileUpdate{
				ID:   -10,
				Name: "valid-name",
			},
		},
		{
			Name:    "failed update: invalid name, too short",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserProfileUpdate{
				ID:   11,
				Name: "",
			},
		},
		{
			Name:    "failed update: not found",
			ResCode: http.StatusNotFound,
			Payload: &model.UserProfileUpdate{
				ID:   createdUserID + 10,
				Name: "valid-name",
			},
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name)
		jsonObject, err := json.Marshal(&tc.Payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		url := appURL + "user-management/" + strconv.Itoa(tc.Payload.ID)
		req, httpReqErr := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal(constants.ShouldNotNil)
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Put("/user-management/:id", userDevController.Update)
		req.Close = true
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer res.Body.Close()
		if res.StatusCode != tc.ResCode {
			t.Error(constants.ShouldEqual, res.StatusCode)
		}
		if tc.Payload != nil {
			respModel := response.Response{}
			decodeErr := json.NewDecoder(res.Body).Decode(&respModel)
			if decodeErr != nil && decodeErr != io.EOF {
				t.Error(constants.ShouldNotErr, decodeErr)
			}
		}
	}
}
