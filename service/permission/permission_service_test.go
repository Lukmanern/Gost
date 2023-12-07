package service

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/errors"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/permission"
	"github.com/stretchr/testify/assert"
)

const (
	fileTestName string = "at PermissionServiceTest"
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
	repository := repository.NewPermissionRepository()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)
	service := NewPermissionService()
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)

	validPermission := createPermission()
	defer repository.Delete(ctx, validPermission.ID)

	type testCase struct {
		Name    string
		Payload model.PermissionCreate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Failed Create Permission -1: name already used",
			Payload: model.PermissionCreate{
				Name:        strings.ToLower(validPermission.Name),
				Description: helper.RandomWords(5),
			},
			WantErr: true,
		},
		{
			Name: "Success Create Permission -1",
			Payload: model.PermissionCreate{
				Name:        strings.ToLower(helper.RandomString(10)),
				Description: helper.RandomWords(5),
			},
			WantErr: false,
		},
		{
			Name: "Success Create Permission -2",
			Payload: model.PermissionCreate{
				Name:        strings.ToLower(helper.RandomString(10)),
				Description: helper.RandomWords(5),
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

		permissionByID, getErr := service.GetByID(ctx, id)
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.Equal(t, permissionByID.Name, tc.Payload.Name, errors.ShouldEqual, tc.Name, fileTestName)
		assert.Equal(t, permissionByID.Description, tc.Payload.Description, errors.ShouldEqual, tc.Name, fileTestName)

		permissionByID, getErr = service.GetByID(ctx, id*99)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, permissionByID, errors.ShouldNil, tc.Name, fileTestName)

		deleteErr := service.Delete(ctx, id)
		assert.NoError(t, deleteErr, errors.ShouldNotErr, fileTestName)

		deleteErr = service.Delete(ctx, id)
		assert.Error(t, deleteErr, errors.ShouldErr, fileTestName)

		permissionByID, getErr = service.GetByID(ctx, id)
		assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
		assert.Nil(t, permissionByID, errors.ShouldNil, tc.Name, fileTestName)
	}
}

func TestGetAll(t *testing.T) {
	repository := repository.NewPermissionRepository()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)
	service := NewPermissionService()
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)

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

		permissions, total, getErr := service.GetAll(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, getErr, errors.ShouldErr, tc.Name, fileTestName)
			continue
		}
		assert.NoError(t, getErr, errors.ShouldNotErr, tc.Name, fileTestName)
		assert.True(t, total >= 0, tc.Name, fileTestName)
		assert.True(t, len(permissions) >= 0, tc.Name, fileTestName)
	}
}

func TestUpdate(t *testing.T) {
	repository := repository.NewPermissionRepository()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)
	service := NewPermissionService()
	assert.NotNil(t, service, errors.ShouldNotNil, fileTestName)

	// create permission for testing {payload}
	permission := createPermission()
	defer repository.Delete(ctx, permission.ID)

	type testCase struct {
		Name    string
		Payload model.PermissionUpdate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: model.PermissionUpdate{
				ID:          permission.ID,
				Name:        strings.ToLower(helper.RandomString(12)),
				Description: helper.RandomWords(7),
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: model.PermissionUpdate{
				ID:          permission.ID,
				Name:        strings.ToLower(helper.RandomString(12)),
				Description: helper.RandomWords(7),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: model.PermissionUpdate{
				ID:   -10,
				Name: strings.ToLower(helper.RandomString(12)),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: data not found",
			Payload: model.PermissionUpdate{
				ID:   permission.ID * 99,
				Name: strings.ToLower(helper.RandomString(12)),
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

		permission, getErr := service.GetByID(ctx, tc.Payload.ID)
		assert.NoError(t, getErr, errors.ShouldNotErr, fileTestName)
		assert.Equal(t, permission.Name, tc.Payload.Name, tc.Name, fileTestName)
		assert.Equal(t, permission.Description, tc.Payload.Description, tc.Name, fileTestName)
	}
}

func createPermission() entity.Permission {
	repository := repository.NewPermissionRepository()
	ctx := helper.NewFiberCtx().Context()
	permissionName := strings.ToLower(helper.RandomString(10))
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
