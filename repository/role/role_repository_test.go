package repository

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
)

var (
	roleRepoImpl RoleRepositoryImpl
	timeNow      time.Time
	ctx          context.Context
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)

	timeNow = time.Now()
	ctx = context.Background()

	roleRepoImpl = RoleRepositoryImpl{
		db: connector.LoadDatabase(),
	}
}

func createOneRole(t *testing.T, namePrefix string) *entity.Role {
	role := entity.Role{
		Name:        "valid-role-name-" + namePrefix,
		Description: "valid-role-description-" + namePrefix,
		TimeFields: entity.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	id, createErr := roleRepoImpl.Create(ctx, role)
	if createErr != nil {
		t.Error("error while creating role : ", createErr.Error())
	}
	role.ID = id
	return &role
}

func TestNewRoleRepository(t *testing.T) {
	roleRepo := NewRoleRepository()
	if roleRepo == nil {
		t.Error("should not nil")
	}
}

func TestCreate(t *testing.T) {
	role := createOneRole(t, "create-same-name")
	if role == nil {
		t.Error("failed creating role : role is nil")
	}
	defer func() {
		roleRepoImpl.Delete(ctx, role.ID)
	}()

	type args struct {
		ctx  context.Context
		role entity.Role
	}
	tests := []struct {
		name      string
		repo      RoleRepositoryImpl
		args      args
		wantErr   bool
		wantPanic bool
	}{
		{
			name:    "error while creating with the same name",
			wantErr: true,
			args: args{
				ctx: ctx,
				role: entity.Role{
					Name:        role.Name,
					Description: "",
					TimeFields: entity.TimeFields{
						CreatedAt: &timeNow,
						UpdatedAt: &timeNow,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil && tt.wantPanic {
					t.Errorf("create() do not panic")
				}
			}()
			gotID, err := tt.repo.Create(tt.args.ctx, tt.args.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotID <= 0 {
				t.Errorf("ID should be positive")
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	role := createOneRole(t, "TestGetByID")
	if role == nil {
		t.Error("failed creating role : role is nil")
	}
	defer func() {
		roleRepoImpl.Delete(ctx, role.ID)
	}()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name: "success get role",
			repo: roleRepoImpl,
			args: args{
				ctx: ctx,
				id:  role.ID,
			},
			wantErr: false,
		},
		{
			name: "failed get role: invalid id",
			repo: roleRepoImpl,
			args: args{
				ctx: ctx,
				id:  -10,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRole, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotRole == nil {
				t.Error("role should not nil")
			}
		})
	}
}

func TestGetByName(t *testing.T) {
	role := createOneRole(t, "TestGetByName")
	if role == nil {
		t.Error("failed creating role : role is nil")
	}
	defer func() {
		roleRepoImpl.Delete(ctx, role.ID)
	}()

	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name: "success get role by valid id",
			repo: roleRepoImpl,
			args: args{
				ctx:  ctx,
				name: role.Name,
			},
			wantErr: false,
		},
		{
			name: "failed get role by invalid id",
			repo: roleRepoImpl,
			args: args{
				ctx:  ctx,
				name: "unknown name",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRole, err := tt.repo.GetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && gotRole == nil {
				t.Error("role should not nil")
			}
		})
	}
}

func TestGetAll(t *testing.T) {
	roles := make([]entity.Role, 0)
	for i := 0; i < 10; i++ {
		role := createOneRole(t, "TestGetAll"+strconv.Itoa(i))
		if role == nil {
			continue
		}
		defer func() {
			roleRepoImpl.Delete(ctx, role.ID)
		}()

		roles = append(roles, *role)
	}
	lenRoles := len(roles)
	type args struct {
		ctx    context.Context
		filter model.RequestGetAll
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name: "success get all",
			repo: roleRepoImpl,
			args: args{
				ctx: ctx,
				filter: model.RequestGetAll{
					Limit: 1000,
					Page:  1,
				},
			},
			wantErr: false,
		},
		{
			name: "success get all",
			repo: roleRepoImpl,
			args: args{
				ctx: ctx,
				filter: model.RequestGetAll{
					Limit: 1,
					Page:  1,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRoles, gotTotal, err := tt.repo.GetAll(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.args.filter.Limit > lenRoles && len(gotRoles) < lenRoles {
				t.Error("role should be $lenRoles or more")
			}
			if tt.args.filter.Limit > lenRoles && gotTotal < lenRoles {
				t.Error("total role should be $lenRoles or more")
			}
			if tt.args.filter.Limit < lenRoles && len(gotRoles) > lenRoles {
				t.Error("role should be less than $lenRoles")
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	role := createOneRole(t, "TestUpdateByID")
	if role == nil {
		t.Error("failed creating role : role is nil")
	}
	defer func() {
		roleRepoImpl.Delete(ctx, role.ID)
	}()

	type args struct {
		ctx  context.Context
		role entity.Role
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name:    "success update name and desc",
			repo:    roleRepoImpl,
			wantErr: false,
			args: args{
				ctx: ctx,
				role: entity.Role{
					ID:          role.ID,
					Name:        "updated name",
					Description: "updated description",
				},
			},
		},
		{
			name:    "failed update name and desc with invalid id",
			repo:    roleRepoImpl,
			wantErr: true,
			args: args{
				ctx: ctx,
				role: entity.Role{
					ID:          -10,
					Name:        "updated name",
					Description: "updated description",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Update(tt.args.ctx, tt.args.role); (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
			}

			p, err := tt.repo.GetByID(tt.args.ctx, role.ID)
			if err != nil {
				t.Error("error while getting role")
			}
			if p.Name != tt.args.role.Name || p.Description != tt.args.role.Description {
				t.Error("name and description failed to update")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	role := createOneRole(t, "TestDeleteByID")
	if role == nil {
		t.Error("failed creating role : role is nil")
	}
	defer func() {
		roleRepoImpl.Delete(ctx, role.ID)
	}()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantErr bool
	}{
		{
			name:    "success update role",
			repo:    roleRepoImpl,
			wantErr: false,
			args: args{
				ctx: ctx,
				id:  role.ID,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			role, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if !tt.wantErr && err == nil {
				t.Error("should error")
			}
			if !tt.wantErr && role != nil {
				t.Error("role should nil")
			}
		})
	}
}
