package controller_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/Lukmanern/gost/internal/response"

	controller "github.com/Lukmanern/gost/controller/user_dev"
	service "github.com/Lukmanern/gost/service/user_dev"
)

var (
	userDevService    service.UserDevService
	userDevController controller.UserDevController
)

func init() {
	// controller\user_dev\user_dev_controller_test.go
	// Check env and database
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	dbURI := config.GetDatabaseURI()
	privKey := config.GetPrivateKey()
	pubKey := config.GetPublicKey()
	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	connector.LoadDatabase()
	r := connector.LoadRedisDatabase()
	r.FlushAll() // clear all key:value in redis

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()

	userDevService = service.NewUserDevService()
	userDevController = controller.NewUserDevController(userDevService)
}

func Test_Create(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userDevController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	createdUser := model.UserCreate{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0],
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
		caseName string
		payload  *model.UserCreate
		wantErr  bool
	}{
		{
			caseName: "success create user -1",
			payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0] + "xyz",
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			wantErr: false,
		},
		{
			caseName: "success create user -2",
			payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0] + "xyz",
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			wantErr: false,
		},
		{
			caseName: "success create user -3",
			payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0] + "xyz",
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			wantErr: false,
		},
		{
			caseName: "failed create user: invalid email address",
			payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    "invalid-email-address",
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			wantErr: true,
		},
		{
			caseName: "failed create user: email already used",
			payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    createdUser.Email,
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			wantErr: true,
		},
		{
			caseName: "failed create user: password too short",
			payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0],
				Password: "short",
				IsAdmin:  true,
			},
			wantErr: true,
		},
		{
			caseName: "failed create user: nil payload, validate failed",
			payload:  nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		jsonObject, marshalErr := json.Marshal(&tc.payload)
		if marshalErr != nil {
			t.Error("should not error", marshalErr.Error())
		}
		c.Request().SetBody(jsonObject)

		createErr := ctr.Create(c)
		if createErr != nil {
			t.Error("should not erro", createErr)
		} else if tc.payload == nil {
			continue
		}

		ctx := c.Context()
		userByEMail, getErr := userDevService.GetByEmail(ctx, tc.payload.Email)
		// if wantErr is false and user is not found
		// there is test failed
		if getErr != nil && !tc.wantErr {
			t.Fatal("test fail", getErr)
		}
		if !tc.wantErr {
			if userByEMail == nil {
				t.Fatal("should not nil")
			} else {
				deleteErr := userDevService.Delete(ctx, userByEMail.ID)
				if deleteErr != nil {
					t.Error("should not error")
				}
			}
			if userByEMail.Name != cases.Title(language.Und).String(tc.payload.Name) {
				t.Error("should equal")
			}
		}
	}
}

func Test_Get(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error("should not nil")
	}

	createdUser := model.UserCreate{
		Name:     helper.RandomString(11),
		Email:    helper.RandomEmails(1)[0],
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
		caseName string
		userID   string
		respCode int
		wantErr  bool
		response response.Response
	}{
		{
			caseName: "success get user",
			userID:   strconv.Itoa(createdUserID),
			respCode: http.StatusOK,
			wantErr:  false,
			response: response.Response{
				Message: response.MessageSuccessLoaded,
				Success: true,
			},
		},
		{
			caseName: "failed get user: negatif user id",
			userID:   "-10",
			respCode: http.StatusBadRequest,
			wantErr:  true,
		},
		{
			caseName: "failed get user: user not found",
			userID:   "199999990",
			respCode: http.StatusNotFound,
			wantErr:  true,
		},
		{
			caseName: "failed get user: failed convert id to int",
			userID:   "not-number",
			respCode: http.StatusBadRequest,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/user-management/"+tc.userID, nil)
		app := fiber.New()
		app.Get("/user-management/:id", userDevController.Get)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal")
		}
		if !tc.wantErr {
			respModel := response.Response{}
			decodeErr := json.NewDecoder(resp.Body).Decode(&respModel)
			if decodeErr != nil {
				t.Error("should not error", decodeErr)
			}

			if tc.response.Message != respModel.Message && tc.response.Message != "" {
				t.Error("should equal")
			}
			if respModel.Success != tc.response.Success {
				t.Error("should equal")
			}
		}
	}
}

func Test_GetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error("should not nil")
	}

	userIDs := make([]int, 0)
	for i := 0; i < 10; i++ {
		createdUser := model.UserCreate{
			Name:     helper.RandomString(11),
			Email:    helper.RandomEmails(1)[0],
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
		caseName string
		payload  string
		respCode int
		wantErr  bool
	}{
		{
			caseName: "success getall",
			payload:  "page=1&limit=100&search=",
			respCode: http.StatusOK,
			wantErr:  false,
		},
		{
			caseName: "failed getall",
			payload:  "page=-1&limit=-100&search=",
			respCode: http.StatusBadRequest,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, "/user-management?"+tc.payload, nil)
		app := fiber.New()
		app.Get("/user-management", userDevController.GetAll)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error", err.Error())
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal")
		}
		if !tc.wantErr {
			body := response.Response{}
			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal("should not error", err.Error())
			}
			err = json.Unmarshal(bytes, &body)
			if err != nil {
				t.Fatal("should not error", err.Error())
			}
			if !body.Success {
				t.Fatal("should be success")
			}
			if len(bytes) <= 2 {
				t.Error("len of bytes should much")
			}
		}
	}
}

func Test_Update(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userDevController
	ctx := c.Context()
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPut)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	createdUser := model.UserCreate{
		Name:     helper.RandomString(11),
		Email:    helper.RandomEmails(1)[0],
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
		caseName string
		payload  *model.UserProfileUpdate
		respCode int
	}{
		{
			caseName: "success update user -1",
			payload: &model.UserProfileUpdate{
				ID:   createdUserID,
				Name: helper.RandomString(6),
			},
			respCode: http.StatusNoContent,
		},
		{
			caseName: "success update user -2",
			payload: &model.UserProfileUpdate{
				ID:   createdUserID,
				Name: helper.RandomString(8),
			},
			respCode: http.StatusNoContent,
		},
		{
			caseName: "success update user -3",
			payload: &model.UserProfileUpdate{
				ID:   createdUserID,
				Name: helper.RandomString(10),
			},
			respCode: http.StatusNoContent,
		},
		{
			caseName: "failed update: invalid id",
			respCode: http.StatusBadRequest,
			payload: &model.UserProfileUpdate{
				ID:   -10,
				Name: "valid-name",
			},
		},
		{
			caseName: "failed update: invalid name, too short",
			respCode: http.StatusBadRequest,
			payload: &model.UserProfileUpdate{
				ID:   11,
				Name: "",
			},
		},
		{
			caseName: "failed update: not found",
			respCode: http.StatusNotFound,
			payload: &model.UserProfileUpdate{
				ID:   createdUserID + 10,
				Name: "valid-name",
			},
		},
	}

	for _, tc := range testCases {
		log.Println(tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := "http://127.0.0.1:9009/user-management/" + strconv.Itoa(tc.payload.ID)
		req, httpReqErr := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Put("/user-management/:id", userDevController.Update)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal", resp.StatusCode)
		}
		if tc.payload != nil {
			respModel := response.Response{}
			decodeErr := json.NewDecoder(resp.Body).Decode(&respModel)
			if decodeErr != nil && decodeErr != io.EOF {
				t.Error("should not error", decodeErr)
			}
		}
	}
}

func Test_Delete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userDevController
	ctx := c.Context()
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPut)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	createdUser := model.UserCreate{
		Name:     helper.RandomString(11),
		Email:    helper.RandomEmails(1)[0],
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
		caseName string
		wantErr  bool
		respCode int
		paramID  int
		response response.Response
	}{
		{
			caseName: "success delete user",
			respCode: http.StatusNoContent,
			paramID:  createdUserID,
		},
		{
			caseName: "failed delete: invalid id",
			respCode: http.StatusBadRequest,
			paramID:  -100,
		},
		{
			caseName: "failed delete: not found",
			respCode: http.StatusNotFound,
			paramID:  createdUserID + 100,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.caseName)
		url := "http://127.0.0.1:9009/user-management/" + strconv.Itoa(tc.paramID)
		req, httpReqErr := http.NewRequest(http.MethodDelete, url, nil)
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Delete("/user-management/:id", userDevController.Delete)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal", resp.StatusCode)
		}
	}

	userByID, err := userDevService.GetByID(ctx, createdUserID)
	if err == nil || userByID != nil {
		t.Error("should error and user should nil")
	}
}
