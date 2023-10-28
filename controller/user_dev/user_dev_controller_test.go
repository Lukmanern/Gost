package controller_test

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	controller "github.com/Lukmanern/gost/controller/user_dev"
	service "github.com/Lukmanern/gost/service/user_dev"
)

var (
	userDevService    service.UserDevService
	userDevController controller.UserController
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
	userDevController = controller.NewUserController(userDevService)
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

		createErr := userDevController.Create(c)
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
	defer func() {
		r := recover()
		if r != nil {
			t.Error("panic ::", r)
		}
	}()

	req := httptest.NewRequest(http.MethodGet, "/user-management/1", nil)
	app := fiber.New()
	app.Get("/user-management/:id", userDevController.Get)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal("should not error")
	}
	if resp == nil {
		t.Error("should not error")
	}
}

// Create(c *fiber.Ctx) error
// Get(c *fiber.Ctx) error
// GetAll(c *fiber.Ctx) error
// Update(c *fiber.Ctx) error
// Delete(c *fiber.Ctx) error
