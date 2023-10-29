package controller

import (
	"log"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	rbacService "github.com/Lukmanern/gost/service/rbac"
	service "github.com/Lukmanern/gost/service/user"
)

var (
	userSvc service.UserService
	userCtr UserController
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

	permService := rbacService.NewPermissionService()
	roleService := rbacService.NewRoleService(permService)
	userSvc = service.NewUserService(roleService)
	userCtr = NewUserController(userSvc)
}

func TestNewUserController(t *testing.T) {
	permService := rbacService.NewPermissionService()
	roleService := rbacService.NewRoleService(permService)
	userService := service.NewUserService(roleService)
	userController := NewUserController(userService)

	if userController == nil || userService == nil || roleService == nil || permService == nil {
		t.Error("should not nil")
	}
}

func Test_Register(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_AccountActivation(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_DeleteAccountActivation(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_ForgetPassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_ResetPassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Login(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Logout(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_UpdatePassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_UpdateProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_MyProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctr := userCtr
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodGet)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}
