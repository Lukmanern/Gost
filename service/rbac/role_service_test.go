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

func TestNewRoleService(t *testing.T) {
	permSvc := NewPermissionService()
	svc := NewRoleService(permSvc)
	if svc == nil {
		t.Error("should not nil")
	}
}

// create 1 role, create 4 permissions
// trying to connect

func TestSuccessCRUD_Role(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	permSvc := NewPermissionService()
	if permSvc == nil || ctx == nil {
		t.Error("should not nil")
	}
	svc := NewRoleService(permSvc)
	if svc == nil {
		t.Error("should not nil")
	}

	// failed create: permissions not found
	func() {
		modelRole := model.RoleCreate{
			Name:          strings.ToLower(helper.RandomString(10)),
			Description:   helper.RandomString(30),
			PermissionsID: []int{-1, -2, -3},
		}
		roleID, createErr := svc.Create(ctx, modelRole)
		if createErr == nil || roleID != 0 {
			t.Error("should error and id should zero")
		}
	}()

	modelRole := model.RoleCreate{
		Name:        strings.ToLower(helper.RandomString(10)),
		Description: helper.RandomString(30),
	}
	roleID, createErr := svc.Create(ctx, modelRole)
	if createErr != nil || roleID < 1 {
		t.Error("should not error and id should more than zero")
	}

	// save the id for delete the perms
	permsID := make([]int, 0)
	for i := 0; i < 3; i++ {
		modelPerm := model.PermissionCreate{
			Name:        strings.ToLower(helper.RandomString(10)),
			Description: helper.RandomString(30),
		}
		permID, createErr := permSvc.Create(ctx, modelPerm)
		if createErr != nil || permID < 1 {
			t.Error("should not error and permID should more than one")
		}

		permsID = append(permsID, permID)
	}

	defer func() {
		svc.Delete(ctx, roleID)
		for _, id := range permsID {
			permSvc.Delete(ctx, id)
		}
	}()

	// failed connect: permissions not
	// found and  role not found
	func() {
		modelConnectFailed := model.RoleConnectToPermissions{
			RoleID:        roleID,
			PermissionsID: []int{-3, -2, -1},
		}
		connectErr := svc.ConnectPermissions(ctx, modelConnectFailed)
		if connectErr == nil {
			t.Error("should error")
		}

		modelConnectFailed = model.RoleConnectToPermissions{
			RoleID:        -1,
			PermissionsID: []int{},
		}
		connectErr = nil
		connectErr = svc.ConnectPermissions(ctx, modelConnectFailed)
		if connectErr == nil {
			t.Error("should error")
		}
	}()

	// success connect
	modelConnect := model.RoleConnectToPermissions{
		RoleID:        roleID,
		PermissionsID: permsID,
	}
	connectErr := svc.ConnectPermissions(ctx, modelConnect)
	if connectErr != nil {
		t.Error("should not error")
	}

	roleByID, getErr := svc.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Error("should not error and role not nil")
	}
	if len(roleByID.Permissions) != len(permsID) {
		t.Error("total of permissions connected by role should equal")
	}

	roles, total, getAllErr := svc.GetAll(ctx, base.RequestGetAll{Limit: 10, Page: 1})
	if len(roles) < 1 || total < 1 || getAllErr != nil {
		t.Error("should more than or equal one and not error at all")
	}

	// update failed : role not found
	func() {
		updateRoleModel := model.RoleUpdate{
			ID:          -1,
			Name:        strings.ToLower(helper.RandomString(11)),
			Description: helper.RandomString(31),
		}
		updateErr := svc.Update(ctx, updateRoleModel)
		if updateErr == nil {
			t.Error("should error")
		}
	}()

	updateRoleModel := model.RoleUpdate{
		ID:          roleID,
		Name:        strings.ToLower(helper.RandomString(11)),
		Description: helper.RandomString(31),
	}
	updateErr := svc.Update(ctx, updateRoleModel)
	if updateErr != nil {
		t.Error("should not error")
	}

	// value reset
	roleByID = nil
	getErr = nil
	roleByID, getErr = svc.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Error("should not error and roleByID should not nil")
	}
	if roleByID.Name != updateRoleModel.Name || roleByID.Description != updateRoleModel.Description {
		t.Error("name and desc should same")
	}

	// delete error : role not found
	func() {
		deleteErr := svc.Delete(ctx, -1)
		if deleteErr == nil {
			t.Error("should error")
		}
	}()

	deleteErr := svc.Delete(ctx, roleID)
	if deleteErr != nil {
		t.Error("should not error")
	}

	// value reset
	roleByID = nil
	getErr = nil
	roleByID, getErr = svc.GetByID(ctx, roleID)
	if getErr == nil || roleByID != nil {
		t.Error("should error and roleByID should nil")
	}
}
