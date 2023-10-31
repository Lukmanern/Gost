package controller

import (
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"

	service "github.com/Lukmanern/gost/service/rbac"
)

var (
	permService2   service.PermissionService
	roleService    service.RoleService
	roleController RoleController
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

	permService2 = service.NewPermissionService()
	roleService = service.NewRoleService(permService2)
	roleController = NewRoleController(roleService)

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()
}

func Test_Role_Create(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
}

func Test_Role_Connect(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
}

func Test_Role_Get(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
}

func Test_Role_GetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
}

func Test_Role_Update(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
}

func Test_Role_Delete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
}
