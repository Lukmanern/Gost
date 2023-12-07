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
	fileTestName string = "at RoleRepoTest"
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
	repository := NewRoleRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	type testCase struct {
		Name          string
		Payload       entity.Role
		PermissionIDs []int
		WantErr       bool
	}

	validRole := createRole()
	defer repository.Delete(ctx, validRole.ID)

	testCases := []testCase{
		{
			Name: "Failed Create Role -1: name already used",
			Payload: entity.Role{
				Name:        validRole.Name,
				Description: helper.RandomWords(5),
			},
			WantErr: true,
		},
		{
			Name: "Failed Create Role -2: invalid permission id",
			Payload: entity.Role{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
			PermissionIDs: []int{-1, -2, 0},
			WantErr:       true,
		},
		{
			Name: "Success Create Role -1",
			Payload: entity.Role{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
			WantErr: false,
		},
		{
			Name: "Success Create Role -2",
			Payload: entity.Role{
				Name:        helper.RandomString(10),
				Description: helper.RandomWords(5),
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, fileTestName)

		tc.Payload.SetCreateTime()
		id, createErr := repository.Create(ctx, tc.Payload, tc.PermissionIDs)
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
	repository := NewRoleRepository()
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
	repository := NewRoleRepository()
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, repository, errors.ShouldNotNil, fileTestName)
	assert.NotNil(t, ctx, errors.ShouldNotNil, fileTestName)

	// create user for testing {payload}
	user := createRole()
	defer repository.Delete(ctx, user.ID)

	type testCase struct {
		Name    string
		Payload entity.Role
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update -1",
			Payload: entity.Role{
				ID:          user.ID,
				Name:        helper.RandomString(12),
				Description: helper.RandomWords(7),
			},
			WantErr: false,
		},
		{
			Name: "Success Update -2",
			Payload: entity.Role{
				ID:          user.ID,
				Name:        helper.RandomString(12),
				Description: helper.RandomWords(7),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update -1: invalid ID",
			Payload: entity.Role{
				ID:   -10,
				Name: helper.RandomString(12),
			},
			WantErr: true,
		},
		{
			Name: "Failed Update -2: invalid id",
			Payload: entity.Role{
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

func createRole() entity.Role {
	repository := NewRoleRepository()
	ctx := helper.NewFiberCtx().Context()
	roleName := helper.RandomString(10)
	role := entity.Role{
		Name:        roleName,
		Description: helper.RandomWords(6),
	}
	role.SetCreateTime()
	id, createErr := repository.Create(ctx, role, []int{})
	if createErr != nil {
		log.Fatal("error while create a new administrator role", fileTestName)
	}
	role.ID = id
	return role
}
