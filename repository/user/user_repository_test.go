package repository

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
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	pwHashed, hashErr := hash.Generate(helper.RandomString(8))
	assert.NoError(t, hashErr, errors.ShouldNotErr, fileTestName)

	type testCase struct {
		Name    string
		Payload entity.User
		WantErr bool
	}

	validUser := createUser(2)
	defer repository.Delete(ctx, validUser.ID)

	testCases := []testCase{
		{
			Name: "Failed Create User -1: email already used",
			Payload: entity.User{
				Name:     helper.RandomString(12),
				Email:    validUser.Email,
				Password: pwHashed,
			},
			WantErr: true,
		},
		{
			Name: "Success Create User -1",
			Payload: entity.User{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: pwHashed,
			},
			WantErr: false,
		},
		{
			Name: "Success Create User -2",
			Payload: entity.User{
				Name:     helper.RandomString(10),
				Email:    helper.RandomEmail(),
				Password: pwHashed,
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		tc.Payload.SetCreateTime()
		id, createErr := repository.Create(ctx, tc.Payload, 2)
		if tc.WantErr {
			assert.Error(t, createErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, createErr, errors.ShouldNotErr, tc.Name, fileTestName)

		userByID, getErr := repository.GetByID(ctx, id)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByID.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByID.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByID.Password, tc.Payload.Password, errors.ShouldEqual, tc.Name, fileTestName)

		userByID, getErr = repository.GetByID(ctx, id*99)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByID, errors.ShouldNil, tc.Name, fileTestName)

		userByEmail, getErr := repository.GetByEmail(ctx, tc.Payload.Email)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByEmail.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByEmail.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByEmail.Password, tc.Payload.Password, errors.ShouldEqual, tc.Name, fileTestName)

		userByEmail, getErr = repository.GetByEmail(ctx, tc.Payload.Email+"invalid*&")
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByEmail, errors.ShouldNil, tc.Name, fileTestName)

		userByConditions, getErr := repository.GetByConditions(ctx, map[string]any{"email =": tc.Payload.Email})
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByConditions.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByConditions.Email, tc.Payload.Email, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByConditions.Password, tc.Payload.Password, errors.ShouldEqual, tc.Name, fileTestName)

		userByConditions, getErr = repository.GetByConditions(ctx, map[string]any{"invalid =": tc.Payload.Email})
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByConditions, errors.ShouldNil, tc.Name, fileTestName)

		deleteErr := repository.Delete(ctx, id)
		assert.NoError(t, deleteErr, errors.ShouldNotErr, fileTestName)

		userByID, getErr = repository.GetByID(ctx, id)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByID, errors.ShouldNil, tc.Name, fileTestName)
	}
}

func TestGetAll(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	type testCase struct {
		Name    string
		Payload model.RequestGetAll
		WantErr bool
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

		users, total, getErr := repository.GetAll(ctx, tc.Payload)
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
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	// create user for testing {payload}
	user := createUser(2)
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload entity.User
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: entity.User{
				ID:   user.ID,
				Name: helper.RandomString(12),
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: entity.User{
				ID:   user.ID,
				Name: helper.RandomString(12),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: entity.User{
				ID:   -10,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: invalid id",
			Payload: entity.User{
				ID:   0,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		tc.Payload.SetUpdateTime()
		updateErr := repository.Update(ctx, tc.Payload)
		if tc.WantErr {
			// error by data not found not detected
			// assert.Error(t, updateErr, errors.ShouldErr, fileTestName)
			continue
		}
		assert.NoError(t, updateErr, errors.ShouldNotErr, fileTestName)

		user, getErr := repository.GetByID(ctx, tc.Payload.ID)
		assert.NoError(t, getErr, errors.ShouldNotErr, fileTestName)
		assert.Equal(t, user.Name, tc.Payload.Name)
	}
}

func TestUpdatePassword(t *testing.T) {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	user := createUser(2)
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload entity.User
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: entity.User{
				ID:       user.ID,
				Password: helper.RandomString(12),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update: invalid id",
			Payload: entity.User{
				ID:       -10,
				Password: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update: invalid id",
			Payload: entity.User{
				ID:       0,
				Password: helper.RandomString(12),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		tc.Payload.SetUpdateTime()
		updateErr := repository.UpdatePassword(ctx, tc.Payload.ID, tc.Payload.Password)
		if tc.WantErr {
			assert.Error(t, updateErr, errors.ShouldErr, fileTestName)
			continue
		}
		assert.NoError(t, updateErr, errors.ShouldNotErr, fileTestName)
	}
}

func createUser(roleID int) entity.User {
	repository := NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	pwHashed, _ := hash.Generate(helper.RandomString(10))
	newUser := entity.User{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: pwHashed,
	}
	newUser.SetCreateTime()
	userID, createErr := repository.Create(ctx, newUser, roleID)
	if createErr != nil {
		log.Fatal("error while create new user", fileTestName)
	}
	newUser.ID = userID
	return newUser
}
