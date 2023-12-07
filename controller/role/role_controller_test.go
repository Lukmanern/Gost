package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/errors"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/response"
	permSvc "github.com/Lukmanern/gost/service/permission"
	service "github.com/Lukmanern/gost/service/role"
)

var (
	permService permSvc.PermissionService
	roleService service.RoleService
	roleCtr     RoleController
	appURL      string
)

func init() {
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appURL = config.AppURL

	connector.LoadDatabase()
	connector.LoadRedisCache()

	permService = permSvc.NewPermissionService()
	roleService = service.NewRoleService(permService)
	roleCtr = NewRoleController(roleService)
}

func TestRoleCreate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(errors.ShouldNotNil)
	}

	totalPermissions := 5
	permIDs := make([]int, 0)
	for i := 0; i < totalPermissions; i++ {
		perm := createPermission(ctx)
		defer func() {
			permService.Delete(ctx, perm.ID)
		}()
		permIDs = append(permIDs, perm.ID)
	}

	role := createRole(ctx, permIDs)
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, role.ID)
	}()
	testCases := []struct {
		caseName string
		respCode int
		payload  model.RoleCreate
	}{
		{
			caseName: "success create -1",
			respCode: http.StatusCreated,
			payload: model.RoleCreate{
				Name:        helper.RandomString(10),
				Description: helper.RandomString(30),
			},
		},
		{
			caseName: "success create -2",
			respCode: http.StatusCreated,
			payload: model.RoleCreate{
				Name:        helper.RandomString(10),
				Description: helper.RandomString(30),
			},
		},
		{
			caseName: "failed create: permissions not found",
			respCode: http.StatusNotFound,
			payload: model.RoleCreate{
				Name:          helper.RandomString(10),
				Description:   helper.RandomString(30),
				PermissionsID: []int{permIDs[0] + 90},
			},
		},
		{
			caseName: "failed create: invalid name, too short",
			respCode: http.StatusBadRequest,
			payload: model.RoleCreate{
				Name:          "",
				Description:   helper.RandomString(30),
				PermissionsID: []int{permIDs[0] - 90},
			},
		},
		{
			caseName: "failed create: invalid description, too short",
			respCode: http.StatusBadRequest,
			payload: model.RoleCreate{
				Name:          helper.RandomString(10),
				Description:   "",
				PermissionsID: []int{permIDs[0]},
			},
		},
	}

	endp := "role"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println("case-name: " + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil {
			t.Error(errors.ShouldNotErr, httpReqErr.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, roleCtr.Create)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(errors.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}

		if resp.StatusCode == http.StatusCreated {
			data, ok1 := data.Data.(map[string]interface{})
			if !ok1 {
				t.Error("should ok1")
			}
			anyID, ok2 := data["id"]
			if !ok2 {
				t.Error("should ok2")
			}
			intID, ok3 := anyID.(float64)
			if !ok3 {
				t.Error("should ok3")
			}
			deleteErr := roleService.Delete(ctx, int(intID))
			if deleteErr != nil {
				t.Error(errors.ShouldNotErr)
			}
		}
	}
}

func TestRoleConnect(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(errors.ShouldNotNil)
	}

	totalPermissions := 5
	permIDs := make([]int, 0)
	for i := 0; i < totalPermissions; i++ {
		perm := createPermission(ctx)
		defer func() {
			permService.Delete(ctx, perm.ID)
		}()
		permIDs = append(permIDs, perm.ID)
	}

	role := createRole(ctx, permIDs)
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, role.ID)
	}()
	roleID := role.ID
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  model.RoleConnectToPermissions
	}{
		{
			caseName: "success connect -1",
			respCode: http.StatusCreated,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "success connect -2",
			respCode: http.StatusCreated,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "failed connect: status not found",
			respCode: http.StatusNotFound,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID + 99,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "failed connect: invalid role id",
			respCode: http.StatusBadRequest,
			payload: model.RoleConnectToPermissions{
				RoleID:        -1,
				PermissionsID: permIDs,
			},
		},
		{
			caseName: "failed connect: invalid id",
			respCode: http.StatusBadRequest,
			payload: model.RoleConnectToPermissions{
				RoleID:        roleID,
				PermissionsID: []int{-1, 2, 3},
			},
		},
	}

	endp := "role/connect"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println("case-name: " + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, roleCtr.Connect)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(errors.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}
		if resp.StatusCode == http.StatusOK {
			role, getErr := roleService.GetByID(ctx, tc.payload.RoleID)
			if getErr != nil || role == nil {
				t.Fatal("should not error while getting role")
			}

			if len(role.Permissions) != len(tc.payload.PermissionsID) {
				t.Error("should equal")
			}
		}
	}
}

func TestRoleGet(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(errors.ShouldNotNil)
	}

	totalPermissions := 5
	permIDs := make([]int, 0)
	for i := 0; i < totalPermissions; i++ {
		perm := createPermission(ctx)
		defer func() {
			permService.Delete(ctx, perm.ID)
		}()
		permIDs = append(permIDs, perm.ID)
	}

	role := createRole(ctx, permIDs)
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, role.ID)
	}()
	roleID := role.ID
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusOK,
			roleID:   roleID,
		},
		{
			caseName: "success get -2",
			respCode: http.StatusOK,
			roleID:   roleID,
		},
		{
			caseName: "failed get: status not found",
			respCode: http.StatusNotFound,
			roleID:   roleID + 99,
		},
		{
			caseName: "failed get: invalid id",
			respCode: http.StatusBadRequest,
			roleID:   -10,
		},
	}

	for _, tc := range testCases {
		url := fmt.Sprintf(appURL+"role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Get("/role/:id", roleCtr.Get)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(errors.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}
	}
}

func TestRoleGetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(errors.ShouldNotNil)
	}

	totalPermissions := 5
	permIDs := make([]int, 0)
	for i := 0; i < totalPermissions; i++ {
		perm := createPermission(ctx)
		defer func() {
			permService.Delete(ctx, perm.ID)
		}()
		permIDs = append(permIDs, perm.ID)
	}

	role := createRole(ctx, permIDs)
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, role.ID)
	}()
	roleID := role.ID
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		params   string
	}{
		{
			caseName: "success getAll -1",
			respCode: http.StatusOK,
			params:   "limit=10&page=1",
		},
		{
			caseName: "success getAll -2",
			respCode: http.StatusOK,
			params:   "limit=100&page=1",
		},
		{
			caseName: "failed getAll: invalid limit/page",
			respCode: http.StatusBadRequest,
			params:   "limit=-10&page=-1",
		},
	}

	for _, tc := range testCases {
		url := appURL + "role?" + tc.params
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Get("/role", roleCtr.GetAll)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(errors.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		var data response.Response
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			t.Fatal("failed to decode JSON:", err)
		}
	}
}

func TestRoleUpdate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(errors.ShouldNotNil)
	}

	totalPermissions := 5
	permIDs := make([]int, 0)
	for i := 0; i < totalPermissions; i++ {
		perm := createPermission(ctx)
		defer func() {
			permService.Delete(ctx, perm.ID)
		}()
		permIDs = append(permIDs, perm.ID)
	}

	role := createRole(ctx, permIDs)
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, role.ID)
	}()
	roleID := role.ID
	if len(role.Permissions) != totalPermissions {
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
		log.Println("case-name: " + tc.caseName)
		jsonObject, err := json.Marshal(tc.payload)
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		url := fmt.Sprintf(appURL+"role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Put("/role/:id", roleCtr.Update)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(errors.ShouldNotErr)
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

func TestRoleDelete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(errors.ShouldNotNil)
	}

	totalPermissions := 5
	permIDs := make([]int, 0)
	for i := 0; i < totalPermissions; i++ {
		perm := createPermission(ctx)
		defer func() {
			permService.Delete(ctx, perm.ID)
		}()
		permIDs = append(permIDs, perm.ID)
	}

	role := createRole(ctx, permIDs)
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, role.ID)
	}()
	roleID := role.ID
	if len(role.Permissions) != totalPermissions {
		t.Error("the length should equal")
	}
	defer func() {
		roleService.Delete(ctx, roleID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		roleID   int
	}{
		{
			caseName: "success update -1",
			respCode: http.StatusNoContent,
			roleID:   roleID,
		},
		{
			caseName: "success update -2",
			respCode: http.StatusNotFound,
			roleID:   roleID,
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
		},
	}

	for _, tc := range testCases {
		url := fmt.Sprintf(appURL+"role/%d", tc.roleID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			t.Error(errors.ShouldNotErr, err.Error())
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Delete("/role/:id", roleCtr.Delete)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(errors.ShouldNotErr)
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
			role, getErr := roleService.GetByID(ctx, tc.roleID)
			if getErr == nil || role != nil {
				t.Error("should error while get role")
			}
		}
	}
}

func createPermission(ctx context.Context) *model.PermissionResponse {
	permID, createErr := permService.Create(ctx, model.PermissionCreate{
		Name:        helper.RandomString(11),
		Description: helper.RandomString(30),
	})
	if createErr != nil || permID < 1 {
		log.Fatal("error while creating permission at Role Controller")
	}
	perm, err := permService.GetByID(ctx, permID)
	if err != nil {
		log.Fatal("error while getting permission at Role Controller")
	}
	return perm
}

func createRole(ctx context.Context, permIDs []int) *entity.Role {
	createdRole := model.RoleCreate{
		Name:          helper.RandomString(9),
		Description:   helper.RandomString(30),
		PermissionsID: permIDs,
	}
	roleID, createErr := roleService.Create(ctx, createdRole)
	if createErr != nil || roleID <= 0 {
		log.Fatal("error while creating new Role at Role Controller")
	}
	roleByID, getErr := roleService.GetByID(ctx, roleID)
	if getErr != nil {
		log.Fatal("error while getting Role at Role Controller")
	}
	return roleByID
}
