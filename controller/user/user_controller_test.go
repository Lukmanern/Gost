package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
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
