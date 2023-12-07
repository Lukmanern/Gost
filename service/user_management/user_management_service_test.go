package service

import (
	"log"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/errors"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/stretchr/testify/assert"
)

const (
	fileTestName string = "at UserRepoTest"
)

var (
	timeNow time.Time
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
}

func TestCreateGetsDelete(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)
	service := NewUserManagementService()
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)

	validUser := createUser(2)
	defer repository.Delete(ctx, validUser.ID)

	type testCase struct {
		Name    string
		Payload model.UserCreate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Create User -1: email already used",
			Payload: model.UserCreate{
				Name:     helper.ToTitle(validUser.Name),
				Email:    validUser.Email,
				Password: "password",
			},
			WantErr: true,
		},
		{
			Name: "Success Create User -1",
			Payload: model.UserCreate{
				Name:     helper.ToTitle(helper.RandomString(10)),
				Email:    helper.RandomEmail(),
				Password: "password",
				IsAdmin:  false,
			},
			WantErr: false,
		},
		{
			Name: "Success Create User -2",
			Payload: model.UserCreate{
				Name:     helper.ToTitle(helper.RandomString(10)),
				Email:    helper.RandomEmail(),
				Password: "password",
				IsAdmin:  true,
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		id, createErr := service.Create(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, createErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, createErr, errors.ShouldNotErr, tc.Name, fileTestName)

		userByID, getErr := service.GetByID(ctx, id)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByID.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByID.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)

		userByEmail, getErr := service.GetByEmail(ctx, tc.Payload.Email)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByEmail.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByEmail.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)

		userByID, getErr = service.GetByID(ctx, id*99)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByID, errors.ShouldNil, tc.Name, fileTestName)

		deleteErr := service.Delete(ctx, id)
		assert.NoError(t, deleteErr, errors.ShouldNotErr, fileTestName)

		deleteErr = service.Delete(ctx, id)
		assert.Error(t, deleteErr, errors.ShouldErr, fileTestName)

		userByID, getErr = service.GetByID(ctx, id)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByID, errors.ShouldNil, tc.Name, fileTestName)
	}
}

func TestGetAll(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)
	service := NewUserManagementService()
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)

	type testCase struct {
		Name    string
		Payload model.RequestGetAll
		WantErr bool
	}

	for i := 0; i < 2; i++ {
		validUser := createUser(2)
		defer repository.Delete(ctx, validUser.ID)
		validUser2 := createUser(2)
		defer repository.Delete(ctx, validUser2.ID)
	}

	testCases := []testCase{
		{
			Name: "Failed Get All -1: invalid sort",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 10,
				Sort:  "invalid-sort",
			},
			WantErr: true,
		},
		{
			Name: "Success Get All -1",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 100,
			},
			WantErr: false,
		},
		{
			Name: "Success Get All -2",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 10,
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		users, total, getErr := service.GetAll(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.True(t, total >= 0, tc.Name, fileTestName)
		assert.True(t, len(users) >= 0, tc.Name, fileTestName)
	}
}

func TestUpdate(t *testing.T) {
	repository := repository.NewUserRepository()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)
	service := NewUserManagementService()
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)

	// create user for testing {payload}
	user := createUser(2)
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload model.UserProfileUpdate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: model.UserProfileUpdate{
				ID:   user.ID,
				Name: helper.ToTitle(helper.RandomString(12)),
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: model.UserProfileUpdate{
				ID:   user.ID,
				Name: helper.ToTitle(helper.RandomString(12)),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: model.UserProfileUpdate{
				ID:   -10,
				Name: helper.ToTitle(helper.RandomString(12)),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: data not found",
			Payload: model.UserProfileUpdate{
				ID:   user.ID * 99,
				Name: helper.ToTitle(helper.RandomString(12)),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		updateErr := service.Update(ctx, tc.Payload)
		if tc.WantErr {
			// error by data not found not detected
			assert.Error(t, updateErr, errors.ShouldErr, fileTestName)
			continue
		}
		assert.NoError(t, updateErr, errors.ShouldNotErr, fileTestName)

		user, getErr := service.GetByID(ctx, tc.Payload.ID)
		assert.NoError(t, getErr, errors.ShouldNotErr, fileTestName)
		assert.Equal(t, user.Name, tc.Payload.Name, tc.Name, fileTestName)
	}
}

func createUser(roleID int) entity.User {
	repository := repository.NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	pwHashed, _ := hash.Generate(helper.RandomString(10))
	user := entity.User{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: pwHashed,
	}
	user.SetCreateTime()
	userID, createErr := repository.Create(ctx, user, roleID)
	if createErr != nil {
		log.Fatal("error while create new user", fileTestName)
	}
	user.ID = userID
	return user
}
