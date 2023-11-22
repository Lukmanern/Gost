package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/constants"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	repository "github.com/Lukmanern/gost/repository/user"
	permService "github.com/Lukmanern/gost/service/permission"
	roleService "github.com/Lukmanern/gost/service/role"
	userService "github.com/Lukmanern/gost/service/user"
	service "github.com/Lukmanern/gost/service/user_management"
)

const (
	testName    = "User Management Controller Test"
	filePath    = "./controller/user_management"
	addTestName = ", at " + testName + " in " + filePath
)

var (
	userSvc           userService.UserService
	userDevService    service.UserManagementService
	userDevController UserManagementController
	userRepo          repository.UserRepository
	appURL            string
)

func init() {
	env.ReadConfig("./../../.env")
	config := env.Configuration()
	appURL = config.AppURL

	connector.LoadDatabase()
	connector.LoadRedisCache()

	userDevService = service.NewUserManagementService()
	userDevController = NewUserManagementController(userDevService)
	userRepo = repository.NewUserRepository()

	permService := permService.NewPermissionService()
	roleService := roleService.NewRoleService(permService)
	userSvc = userService.NewUserService(roleService)
}

func TestCreate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userDevController
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	defer func() {
		userDevService.Delete(c.Context(), createdUser.ID)
		r := recover()
		if r != nil {
			t.Error(addTestName, r)
		}
	}()

	testCases := []struct {
		Name    string
		Payload *model.UserCreate
		ResCode int
	}{
		{
			Name: "success create user -1",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusCreated,
		},
		{
			Name: "success create user -2",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusCreated,
		},
		{
			Name: "success create user -3",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusCreated,
		},
		{
			Name: "failed create user: invalid email address",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    "invalid-email-address",
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusBadRequest,
		},
		{
			Name: "failed create user: email already used",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    createdUser.Email,
				Password: helper.RandomString(11),
				IsAdmin:  true,
			},
			ResCode: fiber.StatusBadRequest,
		},
		{
			Name: "failed create user: password too short",
			Payload: &model.UserCreate{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: "s-;",
				IsAdmin:  true,
			},
			ResCode: fiber.StatusBadRequest,
		},
		{
			Name:    "failed create user: nil Payload, validate failed",
			Payload: nil,
			ResCode: fiber.StatusBadRequest,
		},
	}

	endpoint := "/user-management/"
	for _, tc := range testCases {
		log.Println(tc.Name, addTestName)
		jsonData, marshalErr := json.Marshal(&tc.Payload)
		assert.Nil(t, marshalErr, constants.ShouldNil, marshalErr)

		req, httpReqErr := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(jsonData))
		assert.Nil(t, httpReqErr, constants.ShouldNil, httpReqErr)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		app := fiber.New()
		app.Post(endpoint, ctr.Create)
		req.Close = true

		res, testErr := app.Test(req, -1)
		assert.Nil(t, testErr, constants.ShouldNil, testErr)
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)

		if res.StatusCode == fiber.StatusCreated {
			defer func(email string) {
				u, _ := userRepo.GetByEmail(ctx, email)
				userRepo.Delete(ctx, u.ID)
			}(tc.Payload.Email)
		}
	}
}

func TestGet(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if c == nil || ctx == nil {
		t.Error(constants.ShouldNotNil)
	}

	createdUser := model.UserCreate{
		Name:     helper.RandomString(11),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(11),
		IsAdmin:  true,
	}
	createdUserID, createErr := userDevService.Create(ctx, createdUser)
	if createErr != nil || createdUserID <= 0 {
		t.Error("should not error and more than zero")
	}
	defer func() {
		userDevService.Delete(ctx, createdUserID)
		r := recover()
		if r != nil {
			t.Error("panic ::", r)
		}
	}()

	testCases := []struct {
		caseName string
		userID   string
		respCode int
		wantErr  bool
		response response.Response
	}{
		{
			caseName: "success get user",
			userID:   strconv.Itoa(createdUserID),
			respCode: http.StatusOK,
			wantErr:  false,
			response: response.Response{
				Message: response.MessageSuccessLoaded,
				Success: true,
			},
		},
		{
			caseName: "failed get user: negatif user id",
			userID:   "-10",
			respCode: http.StatusBadRequest,
			wantErr:  true,
		},
		{
			caseName: "failed get user: user not found",
			userID:   "9999",
			respCode: http.StatusNotFound,
			wantErr:  true,
		},
		{
			caseName: "failed get user: failed convert id to int",
			userID:   "not-number",
			respCode: http.StatusBadRequest,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/user-management/"+tc.userID, nil)
		app := fiber.New()
		app.Get("/user-management/:id", userDevController.Get)
		resp, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode != tc.respCode {
			t.Error(constants.ShouldEqual)
		}
		if !tc.wantErr {
			respModel := response.Response{}
			decodeErr := json.NewDecoder(resp.Body).Decode(&respModel)
			if decodeErr != nil {
				t.Error(constants.ShouldNotErr, decodeErr)
			}

			if tc.response.Message != respModel.Message && tc.response.Message != "" {
				t.Error(constants.ShouldEqual)
			}
			if respModel.Success != tc.response.Success {
				t.Error(constants.ShouldEqual)
			}
		}
	}
}

func TestGetAll(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userDevController
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	userIDs := make([]int, 0)
	for i := 0; i < 10; i++ {
		createdUser := createUser(ctx, 1)
		userIDs = append(userIDs, createdUser.ID)
	}

	defer func() {
		for _, id := range userIDs {
			userDevService.Delete(ctx, id)
		}
		r := recover()
		if r != nil {
			t.Error(addTestName, r)
		}
	}()

	testCases := []struct {
		Name    string
		Payload string
		ResCode int
	}{
		{
			Name:    "success getall",
			Payload: "page=1&limit=100&search=",
			ResCode: http.StatusOK,
		},
		{
			Name:    "success getall",
			Payload: "page=2&limit=10&search=",
			ResCode: http.StatusOK,
		},
		{
			Name:    "success getall",
			Payload: "page=3&limit=10&search=",
			ResCode: http.StatusOK,
		},
		{
			Name:    "failed getall",
			Payload: "page=-1&limit=-100&search=",
			ResCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest(http.MethodGet, "/user-management?"+tc.Payload, nil)
		app := fiber.New()
		app.Get("/user-management", userDevController.GetAll)
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr, err.Error())
		}
		defer res.Body.Close()
		assert.Equal(t, res.StatusCode, tc.ResCode, constants.ShouldEqual, res.StatusCode)

		resStruct := response.Response{}
		decodeErr := json.NewDecoder(res.Body).Decode(&resStruct)
		assert.Nil(t, decodeErr, constants.ShouldNil, decodeErr)
	}
}

func TestUpdate(t *testing.T) {
	c := helper.NewFiberCtx()
	ctx := c.Context()
	ctr := userDevController
	assert.NotNil(t, ctr, constants.ShouldNotNil)
	assert.NotNil(t, c, constants.ShouldNotNil)
	assert.NotNil(t, ctx, constants.ShouldNotNil)

	createdUser := createUser(ctx, 1)
	defer func() {
		userDevService.Delete(ctx, createdUser.ID)
		r := recover()
		if r != nil {
			t.Error(addTestName, r)
		}
	}()

	testCases := []struct {
		Name    string
		Payload *model.UserProfileUpdate
		ResCode int
	}{
		{
			Name: "success update user -1",
			Payload: &model.UserProfileUpdate{
				ID:   createdUser.ID,
				Name: helper.RandomString(6),
			},
			ResCode: http.StatusNoContent,
		},
		{
			Name: "success update user -2",
			Payload: &model.UserProfileUpdate{
				ID:   createdUser.ID,
				Name: helper.RandomString(8),
			},
			ResCode: http.StatusNoContent,
		},
		{
			Name: "success update user -3",
			Payload: &model.UserProfileUpdate{
				ID:   createdUser.ID,
				Name: helper.RandomString(10),
			},
			ResCode: http.StatusNoContent,
		},
		{
			Name:    "failed update: invalid id",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserProfileUpdate{
				ID:   -10,
				Name: "valid-name",
			},
		},
		{
			Name:    "failed update: invalid name, too short",
			ResCode: http.StatusBadRequest,
			Payload: &model.UserProfileUpdate{
				ID:   11,
				Name: "",
			},
		},
		{
			Name:    "failed update: not found",
			ResCode: http.StatusNotFound,
			Payload: &model.UserProfileUpdate{
				ID:   createdUser.ID + 10,
				Name: "valid-name",
			},
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name)
		jsonObject, err := json.Marshal(&tc.Payload)
		if err != nil {
			t.Error(constants.ShouldNotErr, err.Error())
		}
		url := appURL + "user-management/" + strconv.Itoa(tc.Payload.ID)
		req, httpReqErr := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonObject))
		if httpReqErr != nil || req == nil {
			t.Fatal(constants.ShouldNotNil)
		}
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		app := fiber.New()
		app.Put("/user-management/:id", userDevController.Update)
		req.Close = true
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal(constants.ShouldNotErr)
		}
		defer res.Body.Close()
		if res.StatusCode != tc.ResCode {
			t.Error(constants.ShouldEqual, res.StatusCode)
		}
		if tc.Payload != nil {
			respModel := response.Response{}
			decodeErr := json.NewDecoder(res.Body).Decode(&respModel)
			if decodeErr != nil && decodeErr != io.EOF {
				t.Error(constants.ShouldNotErr, decodeErr)
			}
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

// func createActiveUser(ctx context.Context, roleID int) (data *entity.User) {
// 	createdUser := model.UserRegister{
// 		Name:     helper.RandomString(10),
// 		Email:    helper.RandomEmail(),
// 		Password: helper.RandomString(10),
// 		RoleID:   1, // admin
// 	}
// 	id, err := userSvc.Register(ctx, createdUser)
// 	if err != nil || id < 1 {
// 		log.Fatal("failed creating user createActiveUser func", err.Error())
// 	}

// 	userByID, getErr := userRepo.GetByID(ctx, id)
// 	if getErr != nil || userByID == nil {
// 		log.Fatal("failed getting user createActiveUser func", getErr.Error())
// 	}
// 	vCode := userByID.VerificationCode
// 	if vCode == nil || userByID.ActivatedAt != nil {
// 		log.Fatal("user should inactivate createActiveUser func"+addTestName, err.Error())
// 	}
// 	userByID.Password = createdUser.Password
// 	verifyErr := userSvc.Verification(ctx, model.UserVerificationCode{
// 		Code:  *vCode,
// 		Email: createdUser.Email,
// 	})
// 	if verifyErr != nil {
// 		log.Fatal("error while user verification createActiveUser func"+addTestName, err.Error())
// 	}

// 	return userByID
// }
