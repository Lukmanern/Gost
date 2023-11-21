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
	"github.com/stretchr/testify/assert"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	repository "github.com/Lukmanern/gost/repository/user"
	permService "github.com/Lukmanern/gost/service/permission"
	roleService "github.com/Lukmanern/gost/service/role"
	service "github.com/Lukmanern/gost/service/user"
)

type testCase struct {
	Name             string
	Handler          func(*fiber.Ctx) error
	AdditionalAction func()
	ResponseCode     int
	Payload          any
}

const (
	testName    = "User Controller Test"
	filePath    = "./controller/user"
	addTestName = ", at " + testName + " in " + filePath
)

var (
	userSvc  service.UserService
	userCtr  UserController
	userRepo repository.UserRepository
	appURL   string
)

func init() {
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appURL = config.AppURL

	connector.LoadDatabase()
	r := connector.LoadRedisCache()
	r.FlushAll() // clear all key:value in redis

	permService := permService.NewPermissionService()
	roleService := roleService.NewRoleService(permService)
	userSvc = service.NewUserService(roleService)
	userCtr = NewUserController(userSvc)
	userRepo = repository.NewUserRepository()
}

func TestNewUserController(t *testing.T) {
	t.Parallel()
	permService := permService.NewPermissionService()
	roleService := roleService.NewRoleService(permService)
	userService := service.NewUserService(roleService)
	userController := NewUserController(userService)

	assert.NotNil(t, userController, constants.ShouldNotNil+addTestName)
	assert.NotNil(t, userService, constants.ShouldNotNil+addTestName)
	assert.NotNil(t, roleService, constants.ShouldNotNil+addTestName)
	assert.NotNil(t, permService, constants.ShouldNotNil+addTestName)
}

func TestRegister(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser, userID := createUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestRegister", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:         "success register -1" + addTestName,
			ResponseCode: http.StatusCreated,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:         "success register -2" + addTestName,
			ResponseCode: http.StatusCreated,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:         "success register -3" + addTestName,
			ResponseCode: http.StatusCreated,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:         "failed register: email already used" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    createdUser.Email,
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:         "failed register: name too short" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserRegister{
				Name:     "",
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:         "failed register: password too short" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: "",
				RoleID:   1, // admin
			},
		},
		{
			Name:         "failed register-1: invalid json body" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload: fiber.Map{
				"body": "false",
			},
		},
		{
			Name:         "failed register-2: invalid json body" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload:      nil,
		},
	}

	endp := "user/register"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.Register)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResponseCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestAccountActivation(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser, userID := createUser(ctx, 1)
	vCode := createdUser.VerificationCode

	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestAccountActivation", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:         "success verify" + addTestName,
			ResponseCode: http.StatusOK,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:         "failed verify: code not found" + addTestName,
			ResponseCode: http.StatusNotFound,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:         "failed verify: code/email too short" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserVerificationCode{
				Code:  "",
				Email: "",
			},
		},
		{
			Name:         "failed verify: invalid json body" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload:      nil,
		},
	}

	endp := "user/verification"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.AccountActivation)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResponseCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestDeleteAccountActivation(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser, userID := createUser(ctx, 1)
	vCode := createdUser.VerificationCode

	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestDeleteAccountActivation", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:         "success delete account" + addTestName,
			ResponseCode: http.StatusOK,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:         "failed delete account: code not found" + addTestName,
			ResponseCode: http.StatusNotFound,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:         "failed delete account: code/email too short" + addTestName,
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserVerificationCode{
				Code:  "",
				Email: "",
			},
		},
	}

	endp := "user/request-delete"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.DeleteAccountActivation)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResponseCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestForgetPassword(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser, userID := createUser(ctx, 1)
	vCode := createdUser.VerificationCode

	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: createdUser.Email,
	})
	if verifyErr != nil {
		t.Fatal("verification should not error")
	}

	// value reset
	createdUser = nil
	createdUser, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || createdUser == nil {
		t.Fatal("getByID should success")
	}
	if createdUser.VerificationCode != nil || createdUser.ActivatedAt == nil {
		t.Fatal("user should active for now, but its get inactive")
	}

	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestForgetPassword ::", r)
		}
	}()

	testCases := []testCase{
		{
			Name:         "success forget password",
			ResponseCode: http.StatusAccepted,
			Payload: &model.UserForgetPassword{
				Email: createdUser.Email,
			},
		},
		{
			Name:         "faield forget password: email not found",
			ResponseCode: http.StatusNotFound,
			Payload: &model.UserForgetPassword{
				Email: helper.RandomEmail(),
			},
		},
		{
			Name:         "faield forget password: invalid email",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserForgetPassword{
				Email: "invalid-email",
			},
		},
	}

	endp := "user/forget-password"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.ForgetPassword)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResponseCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestResetPassword(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	_, userID := createUser(ctx, 1)
	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Fatal("should success get user by id")
	}
	vCode := userByID.VerificationCode
	if vCode == nil || userByID.ActivatedAt != nil {
		t.Fatal("user should inactivate for now, but its get activated/ nulling vCode")
	}
	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: userByID.Email,
	})
	if verifyErr != nil {
		t.Error(constants.ShouldNotErr)
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
		t.Error(constants.ShouldNotErr, "but go err:", forgetPassErr.Error())
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
			t.Fatal("panic at TestResetPassword ::", r)
		}
	}()

	testCases := []testCase{
		{
			Name:         "success reset password",
			ResponseCode: http.StatusAccepted,
			Payload: &model.UserResetPassword{
				Email:              userByID.Email,
				Code:               *userByID.VerificationCode,
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPassword",
			},
		},
		{
			Name:         "failed reset password: password not match",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserResetPassword{
				Email:              userByID.Email,
				Code:               *userByID.VerificationCode,
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPasswordNotMatch",
			},
		},
		{
			Name:         "failed reset password: verification code too short",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserResetPassword{
				Email:              helper.RandomEmail(),
				Code:               "short",
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPasswordNotMatch",
			},
		},
	}

	endp := "user/reset-password"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.ResetPassword)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResponseCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	// create inactive user
	createdUser, userID := createUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestLogin ::", r)
		}
	}()

	// create active user
	createdActiveUser := entity.User{}
	func() {
		createdUser2 := model.UserRegister{
			Name:     helper.RandomString(10),
			Email:    helper.RandomEmail(),
			Password: helper.RandomString(10),
			RoleID:   1, // admin
		}
		userID, createErr := userSvc.Register(ctx, createdUser2)
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

		verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
			Code:  *vCode,
			Email: userByID.Email,
		})
		if verifyErr != nil {
			t.Error(constants.ShouldNotErr)
		}
		userByID = nil
		userByID, getErr = userRepo.GetByID(ctx, userID)
		if getErr != nil || userByID == nil {
			t.Fatal("should success get user by id")
		}

		createdActiveUser = *userByID
		createdActiveUser.Password = createdUser2.Password
	}()

	defer userRepo.Delete(ctx, createdActiveUser.ID)

	testCases := []testCase{
		{
			Name:         "success login",
			ResponseCode: http.StatusOK,
			Payload: &model.UserLogin{
				Email:    createdActiveUser.Email,
				Password: createdActiveUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:         "failed login -1: account is inactive",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Email:    strings.ToLower(createdUser.Email),
				Password: createdUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:         "failed login -2: account is inactive",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Email:    strings.ToLower(createdUser.Email),
				Password: createdUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:         "failed login: wrong passwd",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "wrongPass11",
				Email:    createdUser.Email,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:         "failed login: invalid ip",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "wrongPass11",
				Email:    createdUser.Email,
				IP:       "invalid-ip",
			},
		},
		{
			Name:         "faield login: email not found",
			ResponseCode: http.StatusNotFound,
			Payload: &model.UserLogin{
				Password: "secret123",
				Email:    helper.RandomEmail(),
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:         "faield login: invalid email",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "secret",
				Email:    "invalid-email",
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:         "faield login: Payload too short",
			ResponseCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "",
				Email:    "",
				IP:       helper.RandomIPAddress(),
			},
		},
	}

	endp := "user/login"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.Login)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResponseCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}

	// try blocking IP feature
	clientIP := helper.RandomIPAddress()
	testCase := struct {
		Name         string
		ResponseCode int
		Payload      *model.UserLogin
	}{
		Name:         "failed login: stacking redis",
		ResponseCode: http.StatusBadRequest,
		Payload: &model.UserLogin{
			Email:    createdActiveUser.Email,
			Password: "validpassword",
			IP:       clientIP, // keep the ip same
		},
	}
	for i := 0; i < 7; i++ {
		log.Println("case-name: " + testCase.Name)
		jsonObject, err := json.Marshal(&testCase.Payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		url := appURL + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil {
			t.Fatal(constants.ShouldNotNil)
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.Login)
		req.Close = true
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != testCase.ResponseCode {
			t.Error(testCase.Name, constants.ShouldEqual, resp.StatusCode, "want", testCase.ResponseCode)
		}
	}

	redis := connector.LoadRedisCache()
	if redis == nil {
		t.Fatal(constants.ShouldNotNil)
	}
	value := redis.Get("failed-login-" + clientIP).Val()
	if value != "5" {
		t.Error("should 5, get", value)
	}
}

func TestLogout(t *testing.T) {
	t.Parallel()
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	if ctr == nil || c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	// create inactive user
	createdUser, userID := createUser(ctx, 1)
	vCode := createdUser.VerificationCode

	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: createdUser.Email,
	})
	if verifyErr != nil {
		t.Error(constants.ShouldNotErr)
	}

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	if userToken == "" || loginErr != nil {
		t.Fatal(constants.LoginShouldSuccess)
	}
	defer func() {
		userRepo.Delete(ctx, userID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestLogout ::", r)
		}
	}()

	testCases := []testCase{
		{
			Name:         "success",
			ResponseCode: http.StatusOK,
			Payload:      userToken,
		},
		{
			Name:         "failed logout: fake claims",
			ResponseCode: http.StatusUnauthorized,
			Payload:      "fake-token-123",
		},
		{
			Name:         "failed: payload and token nil",
			ResponseCode: http.StatusUnauthorized,
			Payload:      "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		fakeClaims := jwtHandler.GenerateClaims(tc.Payload.(string))
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.Logout(c)
		resp := c.Response()
		if resp.StatusCode() != tc.ResponseCode {
			t.Error(constants.ShouldEqual, resp.StatusCode())
		}
	}
}

func createUser(ctx context.Context, roleID int) (data *entity.User, id int) {
	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	id, err := userSvc.Register(ctx, createdUser)
	if err != nil || id < 1 {
		log.Fatal("failed creating user at User Controller Test :: createUser func ", err.Error())
	}

	data, getErr := userRepo.GetByID(ctx, id)
	if getErr != nil || data == nil {
		log.Fatal("failed getting user at User Controller Test :: createUser func ", getErr.Error())
	}
	vCode := data.VerificationCode
	if vCode == nil || data.ActivatedAt != nil {
		log.Fatal("user should inactivate at User Controller Test :: createUser func")
	}
	data.Password = createdUser.Password
	return data, id
}

// ForgetPassword(c *fiber.Ctx) error
// ResetPassword(c *fiber.Ctx) error
// Login(c *fiber.Ctx) error
// Logout(c *fiber.Ctx) error
// UpdatePassword(c *fiber.Ctx) error
// UpdateProfile(c *fiber.Ctx) error
// MyProfile(c *fiber.Ctx) error
