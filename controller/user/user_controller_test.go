package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	repository "github.com/Lukmanern/gost/repository/user"
	rbacService "github.com/Lukmanern/gost/service/rbac"
	service "github.com/Lukmanern/gost/service/user"
)

var (
	userSvc  service.UserService
	userCtr  UserController
	userRepo repository.UserRepository
	appUrl   string
)

func init() {
	// controller\user_dev\user_dev_controller_test.go
	// Check env and database
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appUrl = config.AppUrl
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
			t.Fatal("panic ::", r)
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

	endp := "user/register"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
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
			t.Fatal("panic ::", r)
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

	endp := "user/verification"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
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
			t.Fatal("panic ::", r)
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

	endp := "user/request-delete"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
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

	verifyErr := userSvc.Verification(ctx, *vCode)
	if verifyErr != nil {
		t.Fatal("verification should not error")
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, but its get inactive")
	}

	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  *model.UserForgetPassword
	}{
		{
			caseName: "success forget password",
			respCode: http.StatusAccepted,
			payload: &model.UserForgetPassword{
				Email: createdUser.Email,
			},
		},
		{
			caseName: "faield forget password: email not found",
			respCode: http.StatusNotFound,
			payload: &model.UserForgetPassword{
				Email: helper.RandomEmails(1)[0],
			},
		},
		{
			caseName: "faield forget password: invalid email",
			respCode: http.StatusBadRequest,
			payload: &model.UserForgetPassword{
				Email: "invalid-email",
			},
		},
	}

	endp := "user/forget-password"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.ForgetPassword)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode)
		}
	}
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
	verifyErr := userSvc.Verification(ctx, *vCode)
	if verifyErr != nil {
		t.Error("should not error")
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, but its get inactive")
	}

	userForgetPasswd := model.UserForgetPassword{
		Email: userByID.Email,
	}
	forgetPassErr := userSvc.ForgetPassword(ctx, userForgetPasswd)
	if forgetPassErr != nil {
		t.Error("should not error")
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode == nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, and verification code should not nil")
	}

	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		payload  *model.UserResetPassword
	}{
		{
			caseName: "success reset password",
			respCode: http.StatusAccepted,
			payload: &model.UserResetPassword{
				Code:               *userByID.VerificationCode,
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPassword",
			},
		},
		{
			caseName: "failed reset password: password not match",
			respCode: http.StatusBadRequest,
			payload: &model.UserResetPassword{
				Code:               *userByID.VerificationCode,
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPasswordNotMatch",
			},
		},
		{
			caseName: "failed reset password: verification code too short",
			respCode: http.StatusBadRequest,
			payload: &model.UserResetPassword{
				Code:               "short",
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPasswordNotMatch",
			},
		},
	}

	endp := "user/reset-password"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.ResetPassword)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error(tc.caseName, "should equal, but got", resp.StatusCode, "want", tc.respCode)
		}

		if resp.StatusCode == http.StatusAccepted {
			// proofing that password has changed
			token, loginErr := userSvc.Login(ctx, model.UserLogin{
				Email:    userByID.Email,
				Password: tc.payload.NewPassword,
				IP:       helper.RandomIPAddress(),
			})
			if token == "" || loginErr != nil {
				t.Error("should success login, got failed login")
			}
		}
	}
}

func Test_Login(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}
	c.Method(http.MethodPost)
	c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// create inactive user
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
			t.Fatal("panic ::", r)
		}
	}()

	// create active user
	createdActiveUser := entity.User{}
	func() {
		createdUser_2 := model.UserRegister{
			Name:     helper.RandomString(10),
			Email:    helper.RandomEmails(1)[0],
			Password: helper.RandomString(10),
			RoleID:   1, // admin
		}
		userID, createErr := userSvc.Register(ctx, createdUser_2)
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

		verifyErr := userSvc.Verification(ctx, *userByID.VerificationCode)
		if verifyErr != nil {
			t.Error("should not error")
		}
		userByID = nil
		userByID, getErr = userRepo.GetByID(ctx, userID)
		if getErr != nil || userByID == nil {
			t.Fatal("should success get user by id")
		}

		createdActiveUser = *userByID
		createdActiveUser.Password = createdUser_2.Password
	}()

	defer userRepo.Delete(ctx, createdActiveUser.ID)

	testCases := []struct {
		caseName string
		respCode int
		payload  *model.UserLogin
	}{
		{
			caseName: "success login",
			respCode: http.StatusOK,
			payload: &model.UserLogin{
				Email:    createdActiveUser.Email,
				Password: createdActiveUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			caseName: "failed login -1: account is inactive",
			respCode: http.StatusBadRequest,
			payload: &model.UserLogin{
				Email:    strings.ToLower(createdUser.Email),
				Password: createdUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			caseName: "failed login -2: account is inactive",
			respCode: http.StatusBadRequest,
			payload: &model.UserLogin{
				Email:    strings.ToLower(createdUser.Email),
				Password: createdUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			caseName: "failed login: wrong passwd",
			respCode: http.StatusBadRequest,
			payload: &model.UserLogin{
				Password: "wrongPass11",
				Email:    createdUser.Email,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			caseName: "failed login: invalid ip",
			respCode: http.StatusBadRequest,
			payload: &model.UserLogin{
				Password: "wrongPass11",
				Email:    createdUser.Email,
				IP:       "invalid-ip",
			},
		},
		{
			caseName: "faield login: email not found",
			respCode: http.StatusNotFound,
			payload: &model.UserLogin{
				Password: "secret123",
				Email:    helper.RandomEmails(1)[0],
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			caseName: "faield login: invalid email",
			respCode: http.StatusBadRequest,
			payload: &model.UserLogin{
				Password: "secret",
				Email:    "invalid-email",
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			caseName: "faield login: payload too short",
			respCode: http.StatusBadRequest,
			payload: &model.UserLogin{
				Password: "",
				Email:    "",
				IP:       helper.RandomIPAddress(),
			},
		},
	}

	endp := "user/login"
	for _, tc := range testCases {
		log.Println(":::::::" + tc.caseName)
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.Login)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error(tc.caseName, "should equal, but got", resp.StatusCode, "want", tc.respCode)
		}
	}

	// try blocking IP feature
	clientIP := "127.0.0.3"
	testCase := struct {
		caseName string
		respCode int
		payload  *model.UserLogin
	}{
		caseName: "failed login: stacking redis",
		respCode: http.StatusBadRequest,
		payload: &model.UserLogin{
			Email:    createdActiveUser.Email,
			Password: "validpassword",
			IP:       clientIP, // keep the ip same
		},
	}
	for i := 0; i < 7; i++ {
		log.Println(":::::::" + testCase.caseName)
		jsonObject, err := json.Marshal(&testCase.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		url := appUrl + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil {
			t.Fatal("should not nil")
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.Login)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer resp.Body.Close()
		if resp.StatusCode != testCase.respCode {
			t.Error(testCase.caseName, "should equal, but got", resp.StatusCode, "want", testCase.respCode)
		}
	}

	redis := connector.LoadRedisDatabase()
	if redis == nil {
		t.Fatal("should not nil")
	}
	value := redis.Get("failed-login-" + clientIP).Val()
	if value != "5" {
		t.Error("should 5, get", value)
	}
}

func Test_Logout(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	// create inactive user
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

	verifyErr := userSvc.Verification(ctx, *userByID.VerificationCode)
	if verifyErr != nil {
		t.Error("should not error")
	}
	userByID = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, verification code should nil")
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		t.Error("login should success")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		token    string
	}{
		{
			caseName: "success",
			respCode: http.StatusOK,
			token:    userToken,
		},
		{
			caseName: "failed: fake claims",
			respCode: http.StatusUnauthorized,
			token:    "fake-token",
		},
		{
			caseName: "failed: payload nil, token nil",
			respCode: http.StatusUnauthorized,
			token:    "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		fakeClaims := jwtHandler.GenerateClaims(tc.token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.Logout(c)
		resp := c.Response()
		if resp.StatusCode() != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode())
		}

		if resp.StatusCode() == http.StatusOK {
			respBody := c.Response().Body()
			respString := string(respBody)
			respStruct := struct {
				Message string `json:"message"`
				Success bool   `json:"success"`
			}{}

			err := json.Unmarshal([]byte(respString), &respStruct)
			if err != nil {
				t.Errorf("Failed to parse response JSON: %v", err)
			}

			if !respStruct.Success {
				t.Error("Expected success")
			}
		}
	}
}

func Test_UpdatePassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	// create inactive user
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

	verifyErr := userSvc.Verification(ctx, *userByID.VerificationCode)
	if verifyErr != nil {
		t.Error("should not error")
	}
	userByID = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, verification code should nil")
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		t.Error("login should success")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		token    string
		payload  *model.UserPasswordUpdate
	}{
		{
			caseName: "success",
			respCode: http.StatusNoContent,
			token:    userToken,
			payload: &model.UserPasswordUpdate{
				OldPassword:        createdUser.Password,
				NewPassword:        "passwordNew123",
				NewPasswordConfirm: "passwordNew123",
			},
		},
		{
			caseName: "success",
			respCode: http.StatusNoContent,
			token:    userToken,
			payload: &model.UserPasswordUpdate{
				OldPassword:        "passwordNew123",
				NewPassword:        "passwordNew12345",
				NewPasswordConfirm: "passwordNew12345",
			},
		},
		{
			caseName: "failed: no new password",
			respCode: http.StatusBadRequest,
			token:    userToken,
			payload: &model.UserPasswordUpdate{
				OldPassword:        "noNewPassword",
				NewPassword:        "noNewPassword",
				NewPasswordConfirm: "noNewPassword",
			},
		},
		{
			caseName: "failed: payload nil",
			respCode: http.StatusBadRequest,
			token:    userToken,
		},
		{
			caseName: "failed: fake claims",
			respCode: http.StatusUnauthorized,
			token:    "fake-token",
		},
		{
			caseName: "failed: payload nil, token nil",
			respCode: http.StatusUnauthorized,
			token:    "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		if tc.payload != nil {
			requestBody, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatal("Error while serializing payload to request body")
			}
			c.Request().SetBody(requestBody)
		}
		fakeClaims := jwtHandler.GenerateClaims(tc.token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.UpdatePassword(c)
		resp := c.Response()
		if resp.StatusCode() != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode(), "want", tc.respCode)
		}

		if resp.StatusCode() == http.StatusNoContent {
			token, loginErr := userSvc.Login(ctx, model.UserLogin{
				Email:    userByID.Email,
				Password: tc.payload.NewPassword,
				IP:       helper.RandomIPAddress(),
			})
			if loginErr != nil || token == "" {
				t.Error("login should success with new password")
			}
		}
	}
}

func Test_UpdateProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	// create inactive user
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

	verifyErr := userSvc.Verification(ctx, *userByID.VerificationCode)
	if verifyErr != nil {
		t.Error("should not error")
	}
	userByID = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, verification code should nil")
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		t.Error("login should success")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		token    string
		payload  *model.UserProfileUpdate
	}{
		{
			caseName: "success",
			respCode: http.StatusNoContent,
			token:    userToken,
			payload: &model.UserProfileUpdate{
				Name: helper.RandomString(11),
			},
		},
		{
			caseName: "success",
			respCode: http.StatusNoContent,
			token:    userToken,
			payload: &model.UserProfileUpdate{
				Name: helper.RandomString(11),
			},
		},
		{
			caseName: "failed: payload nil",
			respCode: http.StatusBadRequest,
			token:    userToken,
		},
		{
			caseName: "failed: fake claims",
			respCode: http.StatusUnauthorized,
			token:    "fake-token",
		},
		{
			caseName: "failed: payload nil, token nil",
			respCode: http.StatusUnauthorized,
			token:    "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		if tc.payload != nil {
			requestBody, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatal("Error while serializing payload to request body")
			}
			c.Request().SetBody(requestBody)
		}
		fakeClaims := jwtHandler.GenerateClaims(tc.token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.UpdateProfile(c)
		resp := c.Response()
		if resp.StatusCode() != tc.respCode {
			t.Error("should equal, but got", resp.StatusCode(), "want", tc.respCode)
		}

		if resp.StatusCode() == http.StatusNoContent {
			userByID, err := userRepo.GetByID(ctx, userID)
			if err != nil || userByID == nil {
				t.Error("should not error")
			}

			if userByID.Name != cases.Title(language.Und).String(tc.payload.Name) {
				t.Error("shoudl equal")
			}
		}
	}
}

func Test_MyProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error("should not nil")
	}

	// create inactive user
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

	verifyErr := userSvc.Verification(ctx, *userByID.VerificationCode)
	if verifyErr != nil {
		t.Error("should not error")
	}
	userByID = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	if userByID.VerificationCode != nil || userByID.ActivatedAt == nil {
		t.Fatal("user should active for now, verification code should nil")
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		t.Error("login should success")
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		respCode int
		token    string
	}{
		{
			caseName: "success",
			respCode: http.StatusOK,
			token:    userToken,
		},
		{
			caseName: "failed: fake claims",
			respCode: http.StatusUnauthorized,
			token:    "fake-token",
		},
		{
			caseName: "failed: payload nil, token nil",
			respCode: http.StatusUnauthorized,
			token:    "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set("Authorization", fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		fakeClaims := jwtHandler.GenerateClaims(tc.token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.MyProfile(c)
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
			if respStruct.Data.Email != createdUser.Email || respStruct.Data.Role.ID != createdUser.RoleID {
				t.Error("email and other should equal")
			}
		}
	}
}
