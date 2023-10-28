package controller_test

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/gofiber/fiber/v2"

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
	defer func() {
		r := recover()
		if r != nil {
			t.Error("panic ::", r)
		}
	}()
	ctr := userDevController
	if ctr == nil {
		t.Error("should not nil")
	}
	c := helper.NewFiberCtx()
	if ctr == nil || c == nil {
		t.Error("should not error")
	}
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	jsonObject, marshalErr := json.Marshal(&model.UserCreate{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0] + "XYZCOM",
		Password: helper.RandomString(11),
		IsAdmin:  true,
	})
	if marshalErr != nil {
		t.Error("should not error", marshalErr.Error())
	}
	c.Request().SetBody(jsonObject)

	createErr := userDevController.Create(c)
	if createErr != nil {
		t.Error("should not erro", createErr)
	}
}

// Create(c *fiber.Ctx) error
// Get(c *fiber.Ctx) error
// GetAll(c *fiber.Ctx) error
// Update(c *fiber.Ctx) error
// Delete(c *fiber.Ctx) error
