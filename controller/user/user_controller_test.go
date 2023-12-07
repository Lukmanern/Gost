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
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/errors"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	repository "github.com/Lukmanern/gost/repository/user"
	permService "github.com/Lukmanern/gost/service/permission"
	roleService "github.com/Lukmanern/gost/service/role"
	service "github.com/Lukmanern/gost/service/user"
)

type testCase struct {
	Name    string
	Handler func(*fiber.Ctx) error
	ResCode int
	Payload any
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
	permService := permService.NewPermissionService()
	roleService := roleService.NewRoleService(permService)
	userService := service.NewUserService(roleService)
	userController := NewUserController(userService)

	assert.NotNil(t, userController, errors.ShouldNotNil+addTestName)
	assert.NotNil(t, userService, errors.ShouldNotNil+addTestName)
	assert.NotNil(t, roleService, errors.ShouldNotNil+addTestName)
	assert.NotNil(t, permService, errors.ShouldNotNil+addTestName)
}

func TestRegister(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestRegister", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:    "success register -1" + addTestName,
			ResCode: http.StatusCreated,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:    "success register -2" + addTestName,
			ResCode: http.StatusCreated,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:    "success register -3" + addTestName,
			ResCode: http.StatusCreated,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:    "failed register: email already used" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    createdUser.Email,
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:    "failed register: name too short" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: &model.UserRegister{
				Name:     "",
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
				RoleID:   1, // admin
			},
		},
		{
			Name:    "failed register: password too short" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: &model.UserRegister{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: "",
				RoleID:   1, // admin
			},
		},
		{
			Name:    "failed register-1: invalid json body" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: fiber.Map{
				"body": "false",
			},
		},
		{
			Name:    "failed register-2: invalid json body" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: nil,
		},
	}

	endp := "user/register"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, errors.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, errors.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.Register)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, errors.ShouldEqual, res.StatusCode)
		if res.StatusCode == fiber.StatusCreated {
			payload, ok := tc.Payload.(*model.UserRegister)
			if !ok {
				t.Fatal("should ok")
			}
			defer func() {
				u, _ := userRepo.GetByEmail(ctx, payload.Email)
				userRepo.Delete(ctx, u.ID)
			}()
		}

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, errors.ShouldNil, decodeErr)
	}
}

func TestAccountActivation(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	vCode := createdUser.VerificationCode
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestAccountActivation", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:    "success verify" + addTestName,
			ResCode: http.StatusOK,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:    "failed verify: code not found" + addTestName,
			ResCode: http.StatusNotFound,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:    "failed verify: code/email too short" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: &model.UserVerificationCode{
				Code:  "",
				Email: "",
			},
		},
		{
			Name:    "failed verify: invalid json body" + addTestName,
			ResCode: http.StatusBadRequest,
			Payload: nil,
		},
	}

	endp := "user/verification"
	url := appURL + endp
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, errors.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, errors.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.AccountActivation)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, errors.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, errors.ShouldNil, decodeErr)
	}
}

func TestDeleteAccountActivation(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	vCode := createdUser.VerificationCode
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestDeleteAccountActivation", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:    "success delete account" + addTestName,
			ResCode: http.StatusOK,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:    "failed delete account: code not found" + addTestName,
			ResCode: http.StatusNotFound,
			Payload: &model.UserVerificationCode{
				Code:  *vCode,
				Email: createdUser.Email,
			},
		},
		{
			Name:    "failed delete account: code/email too short" + addTestName,
			ResCode: http.StatusBadRequest,
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
		assert.Nil(t, marshalErr, errors.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, errors.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.DeleteAccountActivation)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, errors.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, errors.ShouldNil, decodeErr)
	}
}

func TestForgetPassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	vCode := createdUser.VerificationCode
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestForgetPassword", addTestName, r)
		}
	}()

	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: createdUser.Email,
	})
	assert.Nil(t, verifyErr, "verification should not error")

	createdUser, getErr := userRepo.GetByID(ctx, createdUser.ID)
	assert.Nil(t, getErr, "getByID should succeed")
	assert.NotNil(t, createdUser, "getByID should return a non-nil user")
	assert.Nil(t, createdUser.VerificationCode, "VerificationCode should be nil")
	assert.NotNil(t, createdUser.ActivatedAt, "ActivatedAt should be not nil")

	testCases := []testCase{
		{
			Name:    "success forget password",
			ResCode: http.StatusAccepted,
			Payload: &model.UserForgetPassword{
				Email: createdUser.Email,
			},
		},
		{
			Name:    "faield forget password: email not found",
			ResCode: http.StatusNotFound,
			Payload: &model.UserForgetPassword{
				Email: helper.RandomEmail(),
			},
		},
		{
			Name:    "faield forget password: invalid email",
			ResCode: http.StatusBadRequest,
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
		assert.Nil(t, marshalErr, errors.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, errors.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.ForgetPassword)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, errors.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, errors.ShouldNil, decodeErr)
	}
}

func TestResetPassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestResetPassword", addTestName, r)
		}
	}()

	userByID, getErr := userRepo.GetByID(ctx, createdUser.ID)
	assert.Nil(t, getErr, "Getting user by ID should succeed")
	assert.NotNil(t, userByID, "Getting user by ID should return a non-nil user")
	vCode := userByID.VerificationCode
	assert.NotNil(t, vCode, "VerificationCode should be not nil")
	assert.Nil(t, userByID.ActivatedAt, "ActivatedAt should be nil")

	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: userByID.Email,
	})
	assert.Nil(t, verifyErr, errors.ShouldNotErr)

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, createdUser.ID)
	assert.Nil(t, getErr, "Getting user by ID should succeed")
	assert.NotNil(t, userByID, "Getting user by ID should return a non-nil user")
	assert.Nil(t, userByID.VerificationCode, "VerificationCode should be nil")
	assert.NotNil(t, userByID.ActivatedAt, "ActivatedAt should be nil")

	forgetPassErr := userSvc.ForgetPassword(ctx, model.UserForgetPassword{
		Email: userByID.Email,
	})
	assert.Nil(t, forgetPassErr, errors.ShouldNotErr)

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, createdUser.ID)
	assert.Nil(t, getErr, "Getting user by ID should succeed")
	assert.NotNil(t, userByID, "Getting user by ID should return a non-nil user")
	assert.NotNil(t, userByID.VerificationCode, "VerificationCode should not be nil")
	assert.NotNil(t, userByID.ActivatedAt, "ActivatedAt should not be nil")

	testCases := []testCase{
		{
			Name:    "success reset password",
			ResCode: http.StatusAccepted,
			Payload: &model.UserResetPassword{
				Email:              userByID.Email,
				Code:               *userByID.VerificationCode,
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPassword",
			},
		},
		{
			Name:    "failed reset password: password not match",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserResetPassword{
				Email:              userByID.Email,
				Code:               *userByID.VerificationCode,
				NewPassword:        "newPassword",
				NewPasswordConfirm: "newPasswordNotMatch",
			},
		},
		{
			Name:    "failed reset password: verification code too short",
			ResCode: http.StatusBadRequest,
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
		assert.Nil(t, marshalErr, errors.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, errors.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.ResetPassword)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, errors.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, errors.ShouldNil, decodeErr)
	}
}

func TestLogin(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	// create users
	createdUser := createUser(ctx, 1)
	createdActiveUser := createActiveUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)
		userRepo.Delete(ctx, createdActiveUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestLogin", addTestName, r)
		}
	}()

	testCases := []testCase{
		{
			Name:    "success login",
			ResCode: http.StatusOK,
			Payload: &model.UserLogin{
				Email:    createdActiveUser.Email,
				Password: createdActiveUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:    "failed login -1: account is inactive",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Email:    strings.ToLower(createdUser.Email),
				Password: createdUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:    "failed login -2: account is inactive",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Email:    strings.ToLower(createdUser.Email),
				Password: createdUser.Password,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:    "failed login: wrong passwd",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "wrongPass11",
				Email:    createdUser.Email,
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:    "failed login: invalid ip",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "wrongPass11",
				Email:    createdUser.Email,
				IP:       "invalid-ip",
			},
		},
		{
			Name:    "faield login: email not found",
			ResCode: http.StatusNotFound,
			Payload: &model.UserLogin{
				Password: "secret123",
				Email:    helper.RandomEmail(),
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:    "faield login: invalid email",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserLogin{
				Password: "secret",
				Email:    "invalid-email",
				IP:       helper.RandomIPAddress(),
			},
		},
		{
			Name:    "faield login: Payload too short",
			ResCode: http.StatusBadRequest,
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
		assert.Nil(t, marshalErr, errors.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, errors.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endp, ctr.Login)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, errors.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, errors.ShouldNil, decodeErr)
	}

	// try blocking IP feature
	clientIP := helper.RandomIPAddress()
	testCase := struct {
		Name    string
		ResCode int
		Payload *model.UserLogin
	}{
		Name:    "failed login: stacking redis",
		ResCode: http.StatusBadRequest,
		Payload: &model.UserLogin{
			Email:    createdActiveUser.Email,
			Password: "validpassword",
			IP:       clientIP, // keep the ip same
		},
	}
	for i := 0; i < 10; i++ {
		log.Println(testCase.Name, addTestName)
		jsonObject, marshalErr := json.Marshal(&testCase.Payload)
		assert.Nil(t, marshalErr, errors.ShouldNotErr, addTestName)
		url := appURL + endp
		req, httpReqErr := http.NewRequest(http.MethodPost, url, bytes.NewReader(jsonObject))
		assert.Nil(t, httpReqErr, errors.ShouldNotErr)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Post(endp, ctr.Login)
		req.Close = true
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, errors.ShouldNotErr, addTestName)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, testCase.ResCode, errors.ShouldEqual, res.StatusCode, "want", testCase.ResCode)
	}

	// check value
	redis := connector.LoadRedisCache()
	assert.NotNil(t, redis, errors.ShouldNotNil)
	value := redis.Get("failed-login-" + clientIP).Val()
	assert.Equal(t, value, "5", "value should be 5", addTestName)
}

func TestLogout(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	// create inactive user
	createdUser := createUser(ctx, 1)
	vCode := createdUser.VerificationCode
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestLogout", addTestName, r)
		}
	}()

	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: createdUser.Email,
	})
	assert.NoError(t, verifyErr, "error at user verification TestLogout func", addTestName)

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	assert.NotEmpty(t, userToken, errors.ShouldNotNil, addTestName)
	assert.NoError(t, loginErr, errors.ShouldNotErr, addTestName)

	testCases := []testCase{
		{
			Name:    "success",
			ResCode: http.StatusOK,
			Payload: userToken,
		},
		{
			Name:    "failed logout: fake claims",
			ResCode: http.StatusUnauthorized,
			Payload: "fake-Token-123",
		},
		{
			Name:    "failed: Payload and Token nil",
			ResCode: http.StatusUnauthorized,
			Payload: "",
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
		res := c.Response()
		assert.Equal(t, res.StatusCode(), tc.ResCode)
	}
}

func TestUpdatePassword(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	// create inactive user
	createdUser := createActiveUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestUpdatePassword", addTestName, r)
		}
	}()

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	assert.NotEmpty(t, userToken, errors.ShouldNotNil, addTestName)
	assert.NoError(t, loginErr, errors.ShouldNotErr, addTestName)

	testCases := []struct {
		Name    string
		ResCode int
		Token   string
		Payload *model.UserPasswordUpdate
	}{
		{
			Name:    "success",
			ResCode: http.StatusNoContent,
			Token:   userToken,
			Payload: &model.UserPasswordUpdate{
				OldPassword:        createdUser.Password,
				NewPassword:        "passwordNew123",
				NewPasswordConfirm: "passwordNew123",
			},
		},
		{
			Name:    "success",
			ResCode: http.StatusNoContent,
			Token:   userToken,
			Payload: &model.UserPasswordUpdate{
				OldPassword:        "passwordNew123",
				NewPassword:        "passwordNew12345",
				NewPasswordConfirm: "passwordNew12345",
			},
		},
		{
			Name:    "failed update password: no new password",
			ResCode: http.StatusBadRequest,
			Token:   userToken,
			Payload: &model.UserPasswordUpdate{
				OldPassword:        "noNewPassword",
				NewPassword:        "noNewPassword",
				NewPasswordConfirm: "noNewPassword",
			},
		},
		{
			Name:    "failed update password: Payload nil",
			ResCode: http.StatusBadRequest,
			Token:   userToken,
		},
		{
			Name:    "failed update password: fake claims",
			ResCode: http.StatusUnauthorized,
			Token:   "fake Token 1234",
		},
		{
			Name:    "failed update password: Payload nil, Token nil",
			ResCode: http.StatusUnauthorized,
			Token:   "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		if tc.Payload != nil {
			requestBody, marshalErr := json.Marshal(tc.Payload)
			assert.NoError(t, marshalErr, "Error while serializing Payload to request body", addTestName)
			c.Request().SetBody(requestBody)
		}
		fakeClaims := jwtHandler.GenerateClaims(tc.Token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.UpdatePassword(c)
		res := c.Response()
		assert.Equal(t, res.StatusCode(), tc.ResCode, "want", tc.ResCode, addTestName)
	}
}

func TestUpdateProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	// create inactive user
	createdUser := createActiveUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestUpdateProfile"+addTestName, r)
		}
	}()

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	assert.NotEmpty(t, userToken, errors.ShouldNotNil, addTestName)
	assert.NoError(t, loginErr, errors.ShouldNotErr, addTestName)

	testCases := []struct {
		Name    string
		ResCode int
		Token   string
		Payload *model.UserProfileUpdate
	}{
		{
			Name:    "success",
			ResCode: http.StatusNoContent,
			Token:   userToken,
			Payload: &model.UserProfileUpdate{
				Name: helper.RandomString(11),
			},
		},
		{
			Name:    "success",
			ResCode: http.StatusNoContent,
			Token:   userToken,
			Payload: &model.UserProfileUpdate{
				Name: helper.RandomString(11),
			},
		},
		{
			Name:    "failed: Payload nil",
			ResCode: http.StatusBadRequest,
			Token:   userToken,
		},
		{
			Name:    "failed: fake claims",
			ResCode: http.StatusUnauthorized,
			Token:   "fake-Token",
		},
		{
			Name:    "failed: Payload nil, Token nil",
			ResCode: http.StatusUnauthorized,
			Token:   "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		if tc.Payload != nil {
			requestBody, err := json.Marshal(tc.Payload)
			if err != nil {
				t.Fatal("Error while serializing Payload to request body")
			}
			c.Request().SetBody(requestBody)
		}
		fakeClaims := jwtHandler.GenerateClaims(tc.Token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}
		ctr.UpdateProfile(c)
		res := c.Response()
		assert.Equal(t, res.StatusCode(), tc.ResCode, "want", tc.ResCode, addTestName)
	}
}

func TestMyProfile(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userCtr
	assert.NotNil(t, ctr, errors.ShouldNotNil)
	assert.NotNil(t, c, errors.ShouldNotNil)
	assert.NotNil(t, ctx, errors.ShouldNotNil)

	// create inactive user
	createdUser := createActiveUser(ctx, 1)
	defer func() {
		userRepo.Delete(ctx, createdUser.ID)

		r := recover()
		if r != nil {
			t.Fatal("panic at TestMyProfile"+addTestName, r)
		}
	}()

	userToken, loginErr := userSvc.Login(ctx, model.UserLogin{
		Email:    createdUser.Email,
		Password: createdUser.Password,
		IP:       helper.RandomIPAddress(),
	})
	assert.NotEmpty(t, userToken, errors.ShouldNotNil, addTestName)
	assert.NoError(t, loginErr, errors.ShouldNotErr, addTestName)

	testCases := []struct {
		Name    string
		ResCode int
		Token   string
	}{
		{
			Name:    "success",
			ResCode: http.StatusOK,
			Token:   userToken,
		},
		{
			Name:    "failed: fake claims",
			ResCode: http.StatusUnauthorized,
			Token:   "fake-Token",
		},
		{
			Name:    "failed: payload nil, Token nil",
			ResCode: http.StatusUnauthorized,
			Token:   "",
		},
	}

	jwtHandler := middleware.NewJWTHandler()
	for _, tc := range testCases {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", userToken))
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		c.Request().Header.SetMethod(fiber.MethodGet)
		fakeClaims := jwtHandler.GenerateClaims(tc.Token)
		if fakeClaims != nil {
			c.Locals("claims", fakeClaims)
		}

		ctr.MyProfile(c)
		res := c.Response()
		assert.Equal(t, res.StatusCode(), tc.ResCode, "want", tc.ResCode, addTestName)

		if res.StatusCode() == http.StatusOK {
			resBody := c.Response().Body()
			resString := string(resBody)
			resStruct := response.Response{}
			err := json.Unmarshal([]byte(resString), &resStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func createUser(ctx context.Context, roleID int) (data *entity.User) {
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
	return data
}

func createActiveUser(ctx context.Context, roleID int) (data *entity.User) {
	createdUser := model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(10),
		RoleID:   1, // admin
	}
	id, err := userSvc.Register(ctx, createdUser)
	if err != nil || id < 1 {
		log.Fatal("failed creating user createActiveUser func", err.Error())
	}

	userByID, getErr := userRepo.GetByID(ctx, id)
	if getErr != nil || userByID == nil {
		log.Fatal("failed getting user createActiveUser func", getErr.Error())
	}
	vCode := userByID.VerificationCode
	if vCode == nil || userByID.ActivatedAt != nil {
		log.Fatal("user should inactivate createActiveUser func"+addTestName, err.Error())
	}
	userByID.Password = createdUser.Password
	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
		Code:  *vCode,
		Email: createdUser.Email,
	})
	if verifyErr != nil {
		log.Fatal("error while user verification createActiveUser func"+addTestName, err.Error())
	}

	return userByID
}
