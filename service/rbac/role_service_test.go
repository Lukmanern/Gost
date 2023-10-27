package service

import (
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")
	c := env.Configuration()
	dbURI := c.GetDatabaseURI()
	privKey := c.GetPrivateKey()
	pubKey := c.GetPublicKey()
	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	connector.LoadDatabase()
	connector.LoadRedisDatabase()

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()
}

func TestNewRoleService(t *testing.T) {
	permSvc := NewPermissionService()
	svc := NewRoleService(permSvc)
	if svc == nil {
		t.Error("should not nil")
	}
}

func TestSuccessCRUD_Role(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	svc := NewPermissionService()
	if svc == nil || ctx == nil {
		t.Error("should not nil")
	}
}
