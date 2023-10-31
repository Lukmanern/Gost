package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"

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

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService2.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService2.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
		payload  model.RoleUpdate
	}{
		{
			caseName: "success update -1",
			respCode: http.StatusNoContent,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "success update -2",
			respCode: http.StatusNoContent,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "failed update: invalid name/description",
			respCode: http.StatusBadRequest,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        "",
				Description: "",
			},
		},
		{
			caseName: "failed update: invalid id",
			respCode: http.StatusBadRequest,
			roleID:   -10,
		},
		{
			caseName: "failed update: data not found",
			respCode: http.StatusNotFound,
			roleID:   roleID + 99,
			payload: model.RoleUpdate{
				ID:          roleID + 99,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
	}

	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := fmt.Sprintf("http://127.0.0.1:9009/role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error("should not error", err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Put("/role/:id", roleController.Update)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
		if resp.StatusCode == http.StatusNoContent {
			perm, getErr := roleService.GetByID(ctx, roleID)
			if getErr != nil || perm == nil {
				t.Error("should not nil while get permission")
			}
			if perm.Name != strings.ToLower(tc.payload.Name) ||
				perm.Description != tc.payload.Description {
				t.Error("should equal")
			}
		}
	}
}

func Test_Role_Delete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	permIDs := make([]int, 0)
	for i := 0; i < 4; i++ {
		// create 1 permission
		permID, createErr := permService2.Create(ctx, model.PermissionCreate{
			Name:        helper.RandomString(11),
			Description: helper.RandomString(30),
		})
		if createErr != nil || permID < 1 {
			t.Fatal("should not error while creating permission")
		}
		defer func() {
			permService2.Delete(ctx, permID)
		}()

		permIDs = append(permIDs, permID)
	}

	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		t.Fatal("should not error while creating new Role")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil || roleByID == nil {
		t.Fatal("should not error while getting Role")
	}
	if len(roleByID.Permissions) != 4 {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
		payload  model.RoleUpdate
	}{
		{
			caseName: "success update -1",
			respCode: http.StatusNoContent,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "success update -2",
			respCode: http.StatusNoContent,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
		{
			caseName: "failed update: invalid name/description",
			respCode: http.StatusBadRequest,
			roleID:   roleID,
			payload: model.RoleUpdate{
				ID:          roleID,
				Name:        "",
				Description: "",
			},
		},
		{
			caseName: "failed update: invalid id",
			respCode: http.StatusBadRequest,
			roleID:   -10,
		},
		{
			caseName: "failed update: data not found",
			respCode: http.StatusNotFound,
			roleID:   roleID + 99,
			payload: model.RoleUpdate{
				ID:          roleID + 99,
				Name:        helper.RandomString(12),
				Description: helper.RandomString(20),
			},
		},
	}

	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := fmt.Sprintf("http://127.0.0.1:9009/role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error("should not error", err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Put("/role/:id", roleController.Update)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		if resp.StatusCode != http.StatusNoContent {
			var data response.Response
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				t.Fatal("failed to decode JSON:", err)
			}
		}
		if resp.StatusCode == http.StatusNoContent {
			perm, getErr := roleService.GetByID(ctx, roleID)
			if getErr != nil || perm == nil {
				t.Error("should not nil while get permission")
			}
			if perm.Name != strings.ToLower(tc.payload.Name) ||
				perm.Description != tc.payload.Description {
				t.Error("should equal")
			}
		}
	}
}
