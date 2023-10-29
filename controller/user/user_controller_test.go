package controller

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/response"
	repository "github.com/Lukmanern/gost/repository/user"
	rbacService "github.com/Lukmanern/gost/service/rbac"
	service "github.com/Lukmanern/gost/service/user"
)

var (
	userSvc  service.UserService
	userCtr  UserController
	userRepo repository.UserRepository
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
	userRepo = repository.NewUserRepository()
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
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0],
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	userID, createErr := userSvc.Register(ctx, createdUser)
	if createErr != nil || userID <= 0 {
		t.Fatal("should success create user, user failed to create")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic @ User_Test_Register ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		response response.Response
		payload  *model.UserRegister
	}{
		{
			caseName: "success register -1",
			respCode: http.StatusCreated,
			response: response.Response{
				Message: response.MessageSuccessCreated,
				Success: true,
				Data:    nil,
			},
			payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0],
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			caseName: "success register -2",
			respCode: http.StatusCreated,
			response: response.Response{
				Message: response.MessageSuccessCreated,
				Success: true,
				Data:    nil,
			},
			payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0],
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			caseName: "success register -3",
			respCode: http.StatusCreated,
			response: response.Response{
				Message: response.MessageSuccessCreated,
				Success: true,
				Data:    nil,
			},
			payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0],
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			caseName: "failed register: email already used",
			respCode: http.StatusBadRequest,
			response: response.Response{
				Message: "",
				Success: false,
				Data:    nil,
			},
			payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    createdUser.Email,
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			caseName: "failed register: name too short",
			respCode: http.StatusBadRequest,
			response: response.Response{
				Message: "",
				Success: false,
				Data:    nil,
			},
			payload: &model.UserRegister{
				Name:     "",
				Email:    helper.RandomEmails(1)[0],
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			caseName: "failed register: password too short",
			respCode: http.StatusBadRequest,
			response: response.Response{
				Message: "",
				Success: false,
				Data:    nil,
			},
			payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmails(1)[0],
				Password: "",
				RoleID:   1, // admin
			},
		},
	}

	endp := "/user/register"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := "http://127.0.0.1:9009" + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.Register)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode)
		}

		if tc.payload != nil {
			respModel := response.Response{}
			decodeErr := json.NewDecoder(resp.Body).Decode(&respModel)
			if decodeErr != nil {
				t.Error("should not error", decodeErr)
			}
		}

		if tc.response.Success {
			userByEmail, getErr := userRepo.GetByEmail(ctx, tc.payload.Email)
			if getErr != nil || userByEmail == nil {
				t.Fatal("should success whilte create and get user")
			}
			if userByEmail.Name != cases.Title(language.Und).String(tc.payload.Name) {
				t.Error("name should equal")
			}

			deleteErr := userRepo.Delete(ctx, userByEmail.ID)
			if deleteErr != nil {
				t.Fatal("should success whilte delete user by ID")
			}
		}
	}

}

func Test_AccountActivation(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0],
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	userID, createErr := userSvc.Register(ctx, createdUser)
	if createErr != nil || userID <= 0 {
		t.Fatal("should success create user, user failed to create")
	}
	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	vCode := userByID.VerificationCode
	if vCode == nil || userByID.ActivatedAt != nil {
		t.Fatal("user should inactivate for now, but its get activated/ nulling vCode")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic @ User_Test_Register ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  *model.UserVerificationCode
	}{
		{
			caseName: "success verify",
			respCode: http.StatusOK,
			payload: &model.UserVerificationCode{
				Code: *vCode,
			},
		},
		{
			caseName: "failed verify: code not found",
			respCode: http.StatusNotFound,
			payload: &model.UserVerificationCode{
				Code: *vCode,
			},
		},
		{
			caseName: "failed verify: code too short",
			respCode: http.StatusBadRequest,
			payload: &model.UserVerificationCode{
				Code: "",
			},
		},
	}

	endp := "/user/verification"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := "http://127.0.0.1:9009" + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.AccountActivation)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode)
		}

		// if success
		if resp.StatusCode == http.StatusOK {
			userByEmail, getErr := userRepo.GetByEmail(ctx, createdUser.Email)
			if getErr != nil || userByEmail == nil {
				t.Error("should not error and user not nil")
			}
			if userByEmail.VerificationCode != nil {
				t.Fatal("verif code should nil after activation")
			}
			if userByEmail.ActivatedAt == nil {
				t.Fatal("activated_at should not nil after activation")
			}
		}
	}
}

func Test_DeleteAccountActivation(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0],
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	userID, createErr := userSvc.Register(ctx, createdUser)
	if createErr != nil || userID <= 0 {
		t.Fatal("should success create user, user failed to create")
	}
	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	vCode := userByID.VerificationCode
	if vCode == nil || userByID.ActivatedAt != nil {
		t.Fatal("user should inactivate for now, but its get activated/ nulling vCode")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic @ User_Test_Register ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  *model.UserVerificationCode
	}{
		{
			caseName: "success delete account",
			respCode: http.StatusOK,
			payload: &model.UserVerificationCode{
				Code: *vCode,
			},
		},
		{
			caseName: "failed delete account: code not found",
			respCode: http.StatusNotFound,
			payload: &model.UserVerificationCode{
				Code: *vCode,
			},
		},
		{
			caseName: "failed delete account: code too short",
			respCode: http.StatusBadRequest,
			payload: &model.UserVerificationCode{
				Code: "-",
			},
		},
	}

	endp := "/user/request-delete"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := "http://127.0.0.1:9009" + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.DeleteAccountActivation)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode)
		}

		// if success
		if resp.StatusCode == http.StatusOK {
			userByConds, getErr1 := userRepo.GetByConditions(ctx, map[string]any{
				"verification_code =": tc.payload.Code,
			})
			if getErr1 == nil || userByConds != nil {
				t.Error("should error and user should nil")
			}

			userByID, getErr2 := userRepo.GetByID(ctx, userID)
			if getErr2 == nil || userByID != nil {
				t.Error("should error and user should nil")
			}
		}
	}
}

func Test_ForgetPassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_ResetPassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

// req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
func Test_Login(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_Logout(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_UpdatePassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_UpdateProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}

func Test_MyProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodGet)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
}
