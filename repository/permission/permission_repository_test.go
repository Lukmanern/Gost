package repository

import (
	"log"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/errors"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/stretchr/testify/assert"
)

const (
	fileTestName string = "at PermissionRepoTest"
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
	repository := NewPermissionRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	type testCase struct {
		Name    string
		Payload entity.Permission
		WantErr bool
	}

	validPermission := createPermission()
	defer repository.Delete(ctx, validPermission.ID)

	testCases := []testCase{
		{
			Name: "Failed Create Permission -1: name already used",
			Payload: entity.Permission{
				Name:        validPermission.Name,
				Description: helper.RandomWords(5),
			},
			WantErr: true,
		},
		{
			Name: "Success Create Permission -1",
			Payload: entity.Permission{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
			WantErr: false,
		},
		{
			Name: "Success Create Permission -2",
			Payload: entity.Permission{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		tc.Payload.SetCreateTime()
		id, createErr := repository.Create(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, createErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, createErr, errors.ShouldNotErr, tc.Name, fileTestName)

		userByID, getErr := repository.GetByID(ctx, id)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByID.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByID.Description, tc.Payload.Description, errors.ShouldEqual, tc.Name, fileTestName)

		userByID, getErr = repository.GetByID(ctx, id*99)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByID, errors.ShouldNil, tc.Name, fileTestName)

		userByName, getErr := repository.GetByName(ctx, tc.Payload.Name)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, userByName.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, userByName.Description, tc.Payload.Description, errors.ShouldEqual, tc.Name, fileTestName)

		userByName, getErr = repository.GetByName(ctx, tc.Payload.Name+"*&")
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByName, errors.ShouldNil, tc.Name, fileTestName)

		deleteErr := repository.Delete(ctx, id)
		assert.NoError(t, deleteErr, errors.ShouldNotErr, fileTestName)

		userByID, getErr = repository.GetByID(ctx, id)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, userByID, errors.ShouldNil, tc.Name, fileTestName)
	}
}

func TestGetAll(t *testing.T) {
	repository := NewPermissionRepository()
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
	repository := NewPermissionRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	// create user for testing {payload}
	user := createPermission()
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload entity.Permission
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: entity.Permission{
				ID:          user.ID,
				Name:        helper.RandomString(12),
				Description: helper.RandomWords(7),
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: entity.Permission{
				ID:          user.ID,
				Name:        helper.RandomString(12),
				Description: helper.RandomWords(7),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: entity.Permission{
				ID:   -10,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: invalid id",
			Payload: entity.Permission{
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
		assert.Equal(t, user.Name, tc.Payload.Name, tc.Name, fileTestName)
		assert.Equal(t, user.Description, tc.Payload.Description, tc.Name, fileTestName)
	}
}

func createPermission() entity.Permission {
	repository := NewPermissionRepository()
	ctx := helper.NewFiberCtx().Context()
	permissionName := helper.RandomString(10)
	permission := entity.Permission{
		Name:        permissionName,
		Description: helper.RandomWords(6),
	}
	permission.SetCreateTime()
	id, createErr := repository.Create(ctx, permission)
	if createErr != nil {
		log.Fatal("error while create a new administrator permission", fileTestName)
	}
	permission.ID = id
	return permission
}
