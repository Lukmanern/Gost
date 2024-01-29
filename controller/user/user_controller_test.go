package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/response"
	repository "github.com/Lukmanern/gost/repository/user"
	service "github.com/Lukmanern/gost/service/user"
)

const (
	headerTestName string = "at User Controller Test"
)

var (
	baseURL  string
	timeNow  time.Time
	userRepo repository.UserRepository
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	config := env.Configuration()
	baseURL = config.AppURL
	timeNow = time.Now()
	userRepo = repository.NewUserRepository()

	connector.LoadDatabase()
	r := connector.LoadRedisCache()
	r.FlushAll() // clear all key:value in redis
}

type testCase struct {
	Name    string
	ResCode int
	Payload any
}

func TestUnauthorized(t *testing.T) {
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)

	handlers := []func(c *fiber.Ctx) error{
		controller.GetAll,
		controller.MyProfile,
		controller.Logout,
		controller.UpdateProfile,
		controller.UpdatePassword,
		controller.DeleteAccount,
	}
	for _, handler := range handlers {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		handler(c)
		res := c.Response()
		assert.Equalf(t, res.StatusCode(), fiber.StatusUnauthorized, "Expected response code %d, but got %d", fiber.StatusUnauthorized, res.StatusCode())
	}
}

func TestJSONParser(t *testing.T) {
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	fakeClaims := middleware.Claims{
		Email: helper.RandomEmail(),
		Roles: map[string]uint8{"Full Access": 1},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "999",
			ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(5 * time.Minute)},
			NotBefore: &jwt.NumericDate{Time: time.Now()},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		},
	}

	handlers := []func(c *fiber.Ctx) error{
		controller.Register,
		controller.AccountActivation,
		controller.Login,
		controller.ForgetPassword,
		controller.ResetPassword,
	}
	for _, handler := range handlers {
		c := helper.NewFiberCtx()
		c.Request().Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		c.Locals("claims", &fakeClaims)
		handler(c)
		res := c.Response()
		expectCode := fiber.StatusBadRequest
		assert.Equalf(t, res.StatusCode(), expectCode, "Expected response code %d, but got %d", expectCode, res.StatusCode())
	}
}

func TestRegister(t *testing.T) {
	// Initialize repository, service and controller
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer repository.Delete(ctx, validUser.ID)

	testCases := []testCase{
		{
			Name:    "Success Register -1",
			ResCode: fiber.StatusCreated,
			Payload: model.UserRegister{
				RoleIDs:  []int{1, 2},
				Name:     helper.RandomString(11),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(12),
			},
		},
		{
			Name:    "Success Register -2",
			ResCode: fiber.StatusCreated,
			Payload: model.UserRegister{
				RoleIDs:  []int{1, 2},
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(12),
			},
		},
		{
			Name:    "Failed Register -1: email is already used",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserRegister{
				RoleIDs:  []int{1, 2},
				Name:     validUser.Name,
				Email:    validUser.Email,
				Password: helper.RandomString(12),
			},
		},
		{
			Name:    "Failed Register -2: invalid email",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserRegister{
				RoleIDs:  []int{1, 2},
				Name:     helper.RandomString(10),
				Email:    "invalid email",
				Password: helper.RandomString(12),
			},
		},
		{
			Name:    "Failed Register -3: password too short",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserRegister{
				RoleIDs:  []int{1, 2},
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: "--",
			},
		},
		{
			Name:    "Failed Register -4: no role id",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserRegister{
				RoleIDs:  nil,
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(10),
			},
		},
	}

	pathURL := "user/register"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Marshal payload to JSON
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPost, URL, bytes.NewReader(jsonData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Post(pathURL, controller.Register)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, res.StatusCode)

		if res.StatusCode == fiber.StatusCreated {
			payload, ok := tc.Payload.(model.UserRegister)
			assert.True(t, ok, "should true", headerTestName)
			log.Println(payload)
			entityUser, getErr := repository.GetByEmail(ctx, payload.Email)
			assert.NoError(t, getErr, consts.ShouldNotErr, headerTestName)
			deleteErr := repository.Delete(ctx, entityUser.ID)
			assert.NoError(t, deleteErr, consts.ShouldNotErr, headerTestName)
		}

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestLogin(t *testing.T) {
	// Initialize repository, service and controller
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	entityUser := createUser()
	defer repository.Delete(ctx, entityUser.ID)

	testCases := []testCase{
		{
			Name:    "Success Login -1",
			ResCode: fiber.StatusOK,
			Payload: model.UserLogin{
				Email:    entityUser.Email,
				Password: entityUser.Password,
			},
		},
		{
			Name:    "Success Login -2",
			ResCode: fiber.StatusOK,
			Payload: model.UserLogin{
				Email:    entityUser.Email,
				Password: entityUser.Password,
			},
		},
		{
			Name:    "Failed Login -1 : invalid email",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserLogin{
				Email:    "invalid-email-",
				Password: entityUser.Password,
			},
		},
		{
			Name:    "Failed Login -2 : data not found",
			ResCode: fiber.StatusNotFound,
			Payload: model.UserLogin{
				Email:    "validemail@gost.project",
				Password: entityUser.Password,
			},
		},
		{
			Name:    "Failed Login -3 : password too short",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserLogin{
				Email:    entityUser.Email,
				Password: "--",
			},
		},
	}

	pathURL := "user/login"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Marshal payload to JSON
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPost, URL, bytes.NewReader(jsonData))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Post(pathURL, controller.Login)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, consts.ShouldEqual, res.StatusCode, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestLogout(t *testing.T) {
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)

	tokens := make([]string, 1)
	for i := range tokens {
		tokens[i] = helper.GenerateToken()
	}

	type testCase struct {
		Name    string
		ResCode int
		Token   string
	}

	testCases := []testCase{
		{
			Name:    "Failed Login -1: invalid token",
			ResCode: fiber.StatusUnauthorized,
			Token:   "--",
		},
		{
			Name:    "Failed Login -2: invalid token",
			ResCode: fiber.StatusUnauthorized,
			Token:   "INVALID-TOKEN",
		},
	}

	for i, token := range tokens {
		testCases = append(testCases, testCase{
			Name:    "Success Logout -" + strconv.Itoa(i+2),
			ResCode: fiber.StatusNoContent,
			Token:   token,
		})
		testCases = append(testCases, testCase{
			Name:    "Failed Logout -" + strconv.Itoa(i+3),
			ResCode: fiber.StatusUnauthorized,
			Token:   token,
		})
	}

	pathURL := "user/logout"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPost, URL, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Post(pathURL, jwtHandler.IsAuthenticated, controller.Logout)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, consts.ShouldEqual, res.StatusCode)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestMyProfile(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	tokens := make([]string, 2)
	for i := range tokens {
		tokens[i] = helper.GenerateToken()
	}

	entityUser := createUser()
	validToken, loginErr := service.Login(ctx, model.UserLogin{
		Email:    entityUser.Email,
		Password: entityUser.Password,
	})
	defer repository.Delete(ctx, entityUser.ID)
	assert.NoError(t, loginErr, consts.ShouldNotErr, headerTestName)

	type testCase struct {
		Name    string
		ResCode int
		Token   string
	}

	testCases := []testCase{
		{
			Name:    "Success Get My Profile -1",
			ResCode: fiber.StatusOK,
			Token:   validToken,
		},
		{
			Name:    "Failed Get My Profile -1: invalid token",
			ResCode: fiber.StatusUnauthorized,
			Token:   "--",
		},
		{
			Name:    "Failed Get My Profile -2: invalid token",
			ResCode: fiber.StatusUnauthorized,
			Token:   "INVALID-TOKEN",
		},
	}

	for i, token := range tokens {
		testCases = append(testCases, testCase{
			Name:    "Failed Get My Profile -" + strconv.Itoa(i+3),
			ResCode: fiber.StatusNotFound,
			Token:   token,
		})
	}

	pathURL := "user/my-profile"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodGet, URL, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Get(pathURL, jwtHandler.IsAuthenticated, controller.MyProfile)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, res.StatusCode)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestGetAll(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	token := helper.GenerateToken()
	assert.True(t, token != "", consts.ShouldNotNil, headerTestName)

	type testCase struct {
		Name    string
		Params  string
		ResCode int
		WantErr bool
	}

	testCases := []testCase{
		{
			Name:    "Success get all -1",
			Params:  "?limit=100&page=1",
			ResCode: fiber.StatusOK,
			WantErr: false,
		},
		{
			Name:    "Success get all -2",
			Params:  "?limit=12&page=1",
			ResCode: fiber.StatusOK,
			WantErr: false,
		},
		{
			Name:    "Failed get all: invalid limit",
			Params:  "?limit=-1&page=1",
			ResCode: fiber.StatusBadRequest,
			WantErr: true,
		},
		{
			Name:    "Failed get all: invalid page",
			Params:  "?limit=1&page=-1",
			ResCode: fiber.StatusBadRequest,
			WantErr: true,
		},
		{
			Name:    "Failed get all: invalid sort",
			Params:  "?limit=1&page=1&sort=invalid", // sort should name
			ResCode: fiber.StatusInternalServerError,
			WantErr: true,
		},
	}

	pathURL := "user/"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodGet, URL+tc.Params, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Get(pathURL, jwtHandler.IsAuthenticated, controller.GetAll)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, consts.ShouldEqual, res.StatusCode, res.StatusCode)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestUpdateProfile(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	validUser := createUser()
	defer service.DeleteAccount(ctx, validUser.ID)

	validToken, err := service.Login(ctx, model.UserLogin{
		Email:    validUser.Email,
		Password: validUser.Password,
	})
	assert.NoError(t, err, consts.ShouldNil, err, headerTestName)

	type testCase struct {
		Name    string
		ResCode int
		Payload model.UserUpdate
		Token   string
	}

	testCases := []testCase{
		{
			Name:    "Success Update Profile -1",
			ResCode: fiber.StatusNoContent,
			Payload: model.UserUpdate{
				Name: "test update name",
			},
			Token: validToken,
		},
		{
			Name:    "Failed Update Profile -1: Invalid Token",
			ResCode: fiber.StatusUnauthorized,
			Payload: model.UserUpdate{
				Name: "test update",
			},
			Token: "invalid-token",
		},
		{
			Name:    "Failed Update Profile -2: Name too short",
			ResCode: fiber.StatusBadRequest,
			Payload: model.UserUpdate{
				Name: "",
			},
			Token: helper.GenerateToken(), // valid token
		},
	}

	pathURL := "user/profile"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Marshal payload to JSON
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPut, URL, bytes.NewReader(jsonData))
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Put(pathURL, jwtHandler.IsAuthenticated, controller.UpdateProfile)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr, headerTestName)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, res.StatusCode, tc.Name, headerTestName)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestUpdatePassword(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	fakeToken := helper.GenerateToken()

	entityUser := createUser()
	validToken, loginErr := service.Login(ctx, model.UserLogin{
		Email:    entityUser.Email,
		Password: entityUser.Password,
	})
	defer repository.Delete(ctx, entityUser.ID)
	assert.NoError(t, loginErr, consts.ShouldNotErr, headerTestName)

	type testCase struct {
		Name    string
		ResCode int
		Payload model.UserPasswordUpdate
		Token   string
	}

	testCases := []testCase{
		{
			Name:    "Success Update Password",
			ResCode: fiber.StatusNoContent,
			Token:   validToken,
			Payload: model.UserPasswordUpdate{
				OldPassword:        entityUser.Password,
				NewPassword:        entityUser.Password + "00",
				NewPasswordConfirm: entityUser.Password + "00",
			},
		},
		{
			Name:    "Failed Update Password -1: user not found (invalid token)",
			ResCode: fiber.StatusNotFound,
			Token:   fakeToken,
			Payload: model.UserPasswordUpdate{
				OldPassword:        entityUser.Password,
				NewPassword:        entityUser.Password + "00",
				NewPasswordConfirm: entityUser.Password + "00",
			},
		},
		{
			Name:    "Failed Update Password -2: old and new password is equal",
			ResCode: fiber.StatusBadRequest,
			Token:   validToken,
			Payload: model.UserPasswordUpdate{
				OldPassword:        entityUser.Password,
				NewPassword:        entityUser.Password,
				NewPasswordConfirm: entityUser.Password,
			},
		},
		{
			Name:    "Failed Update Password -3: new and new password confirm is not equal",
			ResCode: fiber.StatusBadRequest,
			Token:   fakeToken,
			Payload: model.UserPasswordUpdate{
				OldPassword:        entityUser.Password,
				NewPassword:        entityUser.Password + "000",
				NewPasswordConfirm: entityUser.Password + "00",
			},
		},
		{
			Name:    "Failed Update Password -4: password too short",
			ResCode: fiber.StatusBadRequest,
			Token:   fakeToken,
			Payload: model.UserPasswordUpdate{
				OldPassword:        "",
				NewPassword:        "" + "000",
				NewPasswordConfirm: "" + "00",
			},
		},
	}

	pathURL := "user/update-password"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Marshal payload to JSON
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.NoError(t, marshalErr, consts.ShouldNotErr, marshalErr)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodPut, URL, bytes.NewReader(jsonData))
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Put(pathURL, jwtHandler.IsAuthenticated, controller.UpdatePassword)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, res.StatusCode)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func TestDeleteAccount(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	service := service.NewUserService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	controller := NewUserController(service)
	assert.NotNil(t, controller, consts.ShouldNotNil, headerTestName)
	jwtHandler := middleware.NewJWTHandler()
	assert.NotNil(t, jwtHandler, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	fakeToken := helper.GenerateToken()

	entityUser := createUser()
	validToken, loginErr := service.Login(ctx, model.UserLogin{
		Email:    entityUser.Email,
		Password: entityUser.Password,
	})
	defer repository.Delete(ctx, entityUser.ID)
	assert.NoError(t, loginErr, consts.ShouldNotErr, headerTestName)

	type testCase struct {
		Name    string
		ResCode int
		Token   string
	}

	testCases := []testCase{
		{
			Name:    "Success Delete Account -1",
			ResCode: fiber.StatusNoContent,
			Token:   validToken,
		},
		{
			Name:    "Failed Delete Account -1: user not found (invalid token)",
			ResCode: fiber.StatusNotFound,
			Token:   fakeToken, // fake but valid
		},
		{
			Name:    "Failed Delete Account -2: user already deleted",
			ResCode: fiber.StatusNotFound,
			Token:   validToken, // is deleted before
		},
	}

	pathURL := "user"
	URL := baseURL + pathURL
	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		// Create HTTP request
		req := httptest.NewRequest(fiber.MethodDelete, URL, nil)
		req.Header.Set(fiber.HeaderAuthorization, fmt.Sprintf("Bearer %s", tc.Token))
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		// Set up Fiber app and handle the request with the controller
		app := fiber.New()
		app.Delete(pathURL, jwtHandler.IsAuthenticated, controller.DeleteAccount)
		req.Close = true

		// run test
		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, consts.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, tc.ResCode, res.StatusCode, consts.ShouldEqual, res.StatusCode)

		if res.StatusCode != fiber.StatusNoContent {
			responseStruct := response.Response{}
			err := json.NewDecoder(res.Body).Decode(&responseStruct)
			assert.NoErrorf(t, err, "Failed to parse response JSON: %v", err)
		}
	}
}

func createUser() entity.User {
	pw := helper.RandomString(15)
	pwHashed, _ := hash.Generate(pw)
	repo := userRepo
	ctx := helper.NewFiberCtx().Context()
	data := entity.User{
		Name:        helper.RandomString(15),
		Email:       helper.RandomEmail(),
		Password:    pwHashed,
		ActivatedAt: &timeNow,
	}
	data.SetCreateTime()
	id, err := repo.Create(ctx, data, []int{1, 2})
	if err != nil {
		log.Fatal("Failed create user", headerTestName)
	}
	data.Password = pw
	data.ID = id
	return data
}
