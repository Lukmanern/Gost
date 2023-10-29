package controller

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

func TestNewDevControllerImpl(t *testing.T) {
	ctr := NewDevControllerImpl()
	c := helper.NewFiberCtx()
	if ctr == nil || c == nil {
		t.Error("should not error")
	}

	pingDbErr := ctr.PingDatabase(c)
	if pingDbErr != nil {
		t.Error("err: ", pingDbErr)
	}

	pingRedisErr := ctr.PingRedis(c)
	if pingRedisErr != nil {
		t.Error("err: ", pingRedisErr)
	}

	panicErr := ctr.Panic(c)
	if panicErr != nil {
		t.Error("err: ", panicErr)
	}

	newJwtErr := ctr.NewJWT(c)
	if newJwtErr != nil {
		t.Error("err: ", newJwtErr)
	}

	// c.Request().Header.Set("Authorization", "Bearer YourJWTToken")
	valJwtErr := ctr.ValidateNewJWT(c)
	if valJwtErr != nil {
		t.Error("err: ", valJwtErr)
	}

	storingErr := ctr.StoringToRedis(c)
	if storingErr != nil {
		t.Error("err: ", storingErr)
	}

	getErr := ctr.GetFromRedis(c)
	if getErr != nil {
		t.Error("err: ", getErr)
	}
}
