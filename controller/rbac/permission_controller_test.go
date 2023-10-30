package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/database/connector"
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
	ctr := permController
	if ctr == nil || c == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Get(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := permController
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

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

		if resp.StatusCode() == http.StatusOK {
			respBody := c.Response().Body()
			respString := string(respBody)
			respStruct := struct {
				Message string            `json:"message"`
				Success bool              `json:"success"`
				Data    model.UserProfile `json:"data"`
			}{}

			err := json.Unmarshal([]byte(respString), &respStruct)
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
