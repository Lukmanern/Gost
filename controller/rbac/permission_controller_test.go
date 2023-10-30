package service

import (
	"log"
	"net/http"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/gofiber/fiber/v2"

	service "github.com/Lukmanern/gost/service/rbac"
)

var (
	permService    service.PermissionService
	permController PermissionController
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
	connector.LoadRedisDatabase()

	permService = service.NewPermissionService()
	permController = NewPermissionController(permService)

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()
}

func Test_NewPermissionController(t *testing.T) {
	permSvc := service.NewPermissionService()
	permCtr := NewPermissionController(permSvc)

	if permSvc == nil || permCtr == nil {
		t.Error("should not nil")
	}
}

func Test_Create(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := permController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Get(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := permController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_GetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := permController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Update(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := permController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Delete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := permController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}
