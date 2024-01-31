package service

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/role"
	"github.com/stretchr/testify/assert"
)

const (
	headerTestName string = "at Role Service Test"
)

var (
	timeNow        time.Time
	roleRepository repository.RoleRepository
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
	roleRepository = repository.NewRoleRepository()
}

// type Role struct {
// 	ID          int    `gorm:"type:serial;primaryKey" json:"id"`
// 	Name        string `gorm:"type:varchar(255) not null unique" json:"name"`
// 	Description string `gorm:"type:varchar(255) not null" json:"description"`
// 	TimeFields
//   }

func TestCreate(t *testing.T) {
	service := NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := roleRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	role := createRole()
	defer repository.Delete(ctx, role.ID)

	type testCase struct {
		Name    string
		Payload model.RoleCreate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Create Role -1",
			Payload: model.RoleCreate{
				Name:        strings.ToLower(helper.RandomString(6)),
				Description: helper.RandomWords(8),
			},
			WantErr: false,
		},
		{
			Name: "Success Create Role -2",
			Payload: model.RoleCreate{
				Name:        strings.ToLower(helper.RandomString(6)),
				Description: helper.RandomWords(8),
			},
			WantErr: false,
		},
		{
			Name: "Failed Create Role -2: name has been used",
			Payload: model.RoleCreate{
				Name:        role.Name,
				Description: helper.RandomWords(8),
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		id, err := service.Create(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)

		// expect no error
		role, getErr := service.GetByID(ctx, id)
		assert.NoError(t, getErr, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.Equal(t, role.Name, tc.Payload.Name, tc.Name, headerTestName)
		assert.Equal(t, role.Description, tc.Payload.Description, tc.Name, headerTestName)

		deleteErr := service.Delete(ctx, id)
		assert.NoError(t, deleteErr, consts.ShouldNotErr, tc.Name, headerTestName)

		// expect error
		_, getErr = service.GetByID(ctx, id)
		assert.Error(t, getErr, consts.ShouldErr, tc.Name, headerTestName)

		deleteErr = service.Delete(ctx, id)
		assert.Error(t, deleteErr, consts.ShouldErr, tc.Name, headerTestName)

	}
}

func TestGetAll(t *testing.T) {
	service := NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := roleRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	totalRoleCreated := 5

	for i := 0; i < totalRoleCreated; i++ {
		role := createRole()
		defer repository.Delete(ctx, role.ID)
	}

	type testCase struct {
		Name    string
		Payload model.RequestGetAll
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Get All Role -1",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 100,
			},
			WantErr: false,
		},
		{
			Name: "Success Get All Role -2",
			Payload: model.RequestGetAll{
				Page:  1,
				Limit: 10 + totalRoleCreated,
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		roles, total, err := service.GetAll(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.True(t, len(roles) >= totalRoleCreated, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.True(t, total >= totalRoleCreated, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func TestUpdate(t *testing.T) {
	service := NewRoleService()
	assert.NotNil(t, service, consts.ShouldNotNil, headerTestName)
	repository := roleRepository
	assert.NotNil(t, repository, consts.ShouldNotNil, headerTestName)
	ctx := helper.NewFiberCtx().Context()
	assert.NotNil(t, ctx, consts.ShouldNotNil, headerTestName)

	role := createRole()
	defer repository.Delete(ctx, role.ID)

	type testCase struct {
		Name    string
		Payload model.RoleUpdate
		WantErr bool
	}

	testCases := []testCase{
		{
			Name: "Success Update Role -1",
			Payload: model.RoleUpdate{
				ID:          role.ID,
				Name:        strings.ToLower(helper.RandomString(5)),
				Description: helper.RandomWords(8),
			},
			WantErr: false,
		},
		{
			Name: "Success Update Role -2",
			Payload: model.RoleUpdate{
				ID:          role.ID,
				Name:        strings.ToLower(helper.RandomString(5)),
				Description: helper.RandomWords(8),
			},
			WantErr: false,
		},
		{
			Name: "Failed Update Role -1: invalid ID",
			Payload: model.RoleUpdate{
				ID: -10,
			},
			WantErr: true,
		},
		{
			Name: "Failed Update Role -2: role not found",
			Payload: model.RoleUpdate{
				ID: role.ID + 999,
			},
			WantErr: true,
		},
		{
			Name: "Failed Update Role -3: name has been used",
			Payload: model.RoleUpdate{
				ID:   role.ID,
				Name: "admin",
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		log.Println(tc.Name, headerTestName)

		err := service.Update(ctx, tc.Payload)
		if tc.WantErr {
			assert.Error(t, err, consts.ShouldErr, tc.Name, headerTestName)
			continue
		}
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)

		role, err := service.GetByID(ctx, tc.Payload.ID)
		assert.NoError(t, err, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.Equal(t, role.Name, tc.Payload.Name, consts.ShouldNotErr, tc.Name, headerTestName)
		assert.Equal(t, role.Description, tc.Payload.Description, consts.ShouldNotErr, tc.Name, headerTestName)
	}
}

func createRole() entity.Role {
	repository := roleRepository
	ctx := helper.NewFiberCtx().Context()
	role := entity.Role{
		Name:        strings.ToLower(helper.RandomString(15)),
		Description: helper.RandomWords(8),
	}
	role.SetCreateTime()
	id, err := repository.Create(ctx, role)
	if err != nil {
		log.Fatal("failed create a new user", headerTestName)
	}
	role.ID = id
	return role
}
