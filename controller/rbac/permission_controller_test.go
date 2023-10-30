package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/Lukmanern/gost/internal/response"

	userController "github.com/Lukmanern/gost/controller/user"
	userRepository "github.com/Lukmanern/gost/repository/user"
	service "github.com/Lukmanern/gost/service/rbac"
	userService "github.com/Lukmanern/gost/service/user"
)

var (
	userRepo       userRepository.UserRepository
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

	userRepo = userRepository.NewUserRepository()
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
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	userID, userToken := createUserAndToken()
	if userID < 1 || len(userToken) < 2 {
		t.Error("should more")
	}

	defer func() {
		userRepo.Delete(ctx, userID)
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  model.PermissionCreate
		token    string // jwt into claims (fake claims)
	}{
		{
			caseName: "success create",
			respCode: http.StatusCreated,
			payload: model.PermissionCreate{
				Name:        "example-permission-001",
				Description: "example-description-of-permission-001",
			},
			token: userToken,
		},
		{
			caseName: "failed create: name already used",
			respCode: http.StatusBadRequest,
			payload: model.PermissionCreate{
				Name:        "example-permission-001",
				Description: "example-description-of-permission-001",
			},
			token: userToken,
		},
		{
			caseName: "failed create: name/desc too short",
			respCode: http.StatusBadRequest,
			payload:  model.PermissionCreate{},
			token:    userToken,
		},
	}

	createdIDs := make([]float64, 0)
	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", tc.token))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		requestBody, err := json.Marshal(tc.payload)
		if err != nil {
			t.Fatal("Error while serializing payload to request body")
		}
		c.Request().SetBody(requestBody)

		fakeClaims := jwtHandler.GenerateClaims(tc.token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.Create(c)
		resp := c.Response()
		if resp.StatusCode() != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode())
		}

		if resp.StatusCode() == http.StatusCreated {
			respBody := c.Response().Body()
			respString := string(respBody)
			respStruct := struct {
				Message string         `json:"message"`
				Success bool           `json:"success"`
				Data    map[string]any `json:"data"`
			}{}

			err := json.Unmarshal([]byte(respString), &respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}
			if !respStruct.Success {
				t.Error("Expected success")
			}
			if respStruct.Message != response.MessageSuccessCreated {
				t.Error("Expected message to be equal")
			}
			if id, ok := respStruct.Data["id"].(float64); !ok || id < 1 {
				t.Error("should be a positive integer")
			} else {
				createdIDs = append(createdIDs, id)
			}
		}
	}

	for _, id := range createdIDs {
		err := permService.Delete(ctx, int(id))
		if err != nil {
			t.Error("deletingc created permission/s should not error")
		}
	}
}

func Test_Get(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	testCases := []struct {
		caseName string
		respCode int
		permID   int
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusOK,
			permID:   1,
		},
		{
			caseName: "success get -2",
			respCode: http.StatusOK,
			permID:   1,
		},
		{
			caseName: "failed get: invalid id",
			respCode: http.StatusBadRequest,
			permID:   -10,
		},
		{
			caseName: "failed get: data not found",
			respCode: http.StatusNotFound,
			permID:   99999,
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/permission/%d", tc.permID), nil)
		app := fiber.New()
		app.Get("/permission/:id", permController.Get)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, want", tc.respCode, "but got", resp.StatusCode)
		}
		if resp.StatusCode == http.StatusOK {
			var respStruct struct {
				Message string              `json:"message"`
				Success bool                `json:"success"`
				Data    base.GetAllResponse `json:"data"`
			}
			err := json.NewDecoder(resp.Body).Decode(&respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}
			if !respStruct.Success {
				t.Error("Expected success")
			}
			if respStruct.Message != response.MessageSuccessLoaded {
				t.Error("Expected message to be equal")
			}
		}
	}
}

func Test_GetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	testCases := []struct {
		caseName string
		respCode int
		payload  base.RequestGetAll
	}{
		{
			caseName: "success get -1",
			respCode: http.StatusOK,
			payload: base.RequestGetAll{
				Limit: 10,
				Page:  1,
			},
		},
		{
			caseName: "success get -2",
			respCode: http.StatusOK,
			payload: base.RequestGetAll{
				Limit: 100,
				Page:  2,
			},
		},
		{
			caseName: "failed get: invalid payload",
			respCode: http.StatusBadRequest,
			payload: base.RequestGetAll{
				Limit: -1,
				Page:  -1,
			},
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/permission?page=%d&limit=%d", tc.payload.Page, tc.payload.Limit), nil)
		app := fiber.New()
		app.Get("/permission", permController.GetAll)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal")
		}
		if resp.StatusCode == http.StatusOK {
			var respStruct struct {
				Message string              `json:"message"`
				Success bool                `json:"success"`
				Data    base.GetAllResponse `json:"data"`
			}
			err := json.NewDecoder(resp.Body).Decode(&respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}
			if !respStruct.Success {
				t.Error("Expected success")
			}
			if respStruct.Message != response.MessageSuccessLoaded {
				t.Error("Expected message to be equal")
			}
			if len(respStruct.Data.Data.([]any)) > tc.payload.Limit {
				t.Error("should less or equal", len(respStruct.Data.Data.([]any)))
			}
		}
	}
}

func Test_Update(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

}

func Test_Delete(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

}

func createUserAndToken() (userID int, token string) {
	permService := service.NewPermissionService()
	roleService := service.NewRoleService(permService)
	userSvc := userService.NewUserService(roleService)
	userCtr := userController.NewUserController(userSvc)

	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		log.Fatal("should not nil")
	}

	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0],
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	userID, createErr := userSvc.Register(ctx, createdUser)
	if createErr != nil || userID <= 0 {
		log.Fatal("should success create user, user failed to create")
	}
	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		log.Fatal("should success get user by id")
	}
	vCode := userByID.VerificationCode
	if vCode == nil || userByID.ActivatedAt != nil {
		log.Fatal("user should inactivate for now, but its get activated/ nulling vCode")
	}

	verifyErr := userSvc.Verification(ctx, *userByID.VerificationCode)
	if verifyErr != nil {
		log.Fatal("should not error")
	}
	userByID = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		log.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		log.Fatal("user should active for now, verification code should nil")
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		log.Fatal("login should success")
	}

	return userID, userToken
}
