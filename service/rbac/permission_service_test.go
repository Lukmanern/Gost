package service

import (
	"log"
	"strings"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/model"
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

func TestNewPermissionService(t *testing.T) {
	svc := NewPermissionService()
	if svc == nil {
		t.Error("should not nil")
	}
}

// Create 1 role
// -> get by id
// -> get all and check >= 1
// -> update
// -> delete
// -> get by id

func TestSuccessCRUD_Permission(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	svc := NewPermissionService()
	if svc == nil || ctx == nil {
		t.Error("should not nil")
	}
	modelPerm := model.PermissionCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	permID, createErr := svc.Create(ctx, modelPerm)
	if createErr != nil || permID < 1 {
		t.Error("should not error and permID should more than one")
	}
	defer func() {
		svc.Delete(ctx, permID)
	}()

	permByID, getErr := svc.GetByID(ctx, permID)
	if getErr != nil || permByID == nil {
		t.Error("should not error and permByID should not nil")
	}
	if permByID.Name != modelPerm.Name || permByID.Description != modelPerm.Description {
		t.Error("name and desc should same")
	}

	perms, total, getAllErr := svc.GetAll(ctx, base.RequestGetAll{Limit: 10, Page: 1})
	if len(perms) < 1 || total < 1 || getAllErr != nil {
		t.Error("should more than or equal one and not error at all")
	}

	updatePermModel := model.PermissionUpdate{
		ID:          permID,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updatePermModel)
	if updateErr != nil {
		t.Error("should not error")
	}

	// value reset
	permByID = nil
	getErr = nil
	permByID, getErr = svc.GetByID(ctx, permID)
	if getErr != nil || permByID == nil {
		t.Error("should not error and permByID should not nil")
	}
	if permByID.Name != updatePermModel.Name || permByID.Description != updatePermModel.Description {
		t.Error("name and desc should same")
	}

	deleteErr := svc.Delete(ctx, permID)
	if deleteErr != nil {
		t.Error("should not error")
	}

	// value reset
	permByID = nil
	getErr = nil
	permByID, getErr = svc.GetByID(ctx, permID)
	if getErr == nil || permByID != nil {
		t.Error("should error and permByID should nil")
	}
}

// Create(ctx context.Context, permission model.PermissionCreate) (id int, err error)
// GetByID(ctx context.Context, id int) (permission *model.PermissionResponse, err error)
// GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []model.PermissionResponse, total int, err error)
// Update(ctx context.Context, permission model.PermissionUpdate) (err error)
// Delete(ctx context.Context, id int) (err error)

func TestFailedCRUD_Permission(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	svc := NewPermissionService()
	if svc == nil || ctx == nil {
		t.Error("should not nil")
	}
	modelPerm := model.PermissionCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	permID, createErr := svc.Create(ctx, modelPerm)
	if createErr != nil || permID < 1 {
		t.Error("should not error and permID should more than one")
	}
	defer func() {
		svc.Delete(ctx, permID)
	}()

	permByID, getErr := svc.GetByID(ctx, -10)
	if getErr == nil || permByID != nil {
		t.Error("should error and permByID should nil")
	}

	updatePermModel := model.PermissionUpdate{
		ID:          -10,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updatePermModel)
	if updateErr == nil {
		t.Error("should error")
	}

	deleteErr := svc.Delete(ctx, -10)
	if deleteErr == nil {
		t.Error("should error")
	}
}
