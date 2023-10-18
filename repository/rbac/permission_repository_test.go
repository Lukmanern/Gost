package repository

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/env"
)

var (
	permissionRepoImpl PermissionRepositoryImpl
	timeNow            time.Time
	ctx                context.Context
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)
	timeNow = time.Now()
	ctx = context.Background()
	permissionRepoImpl = PermissionRepositoryImpl{
		permissionTableName: permissionTableName,
		db:                  connector.LoadDatabase(),
	}

}

func createOnePermission(t *testing.T, namePrefix string) *entity.Permission {
	// create permission
	permission := entity.Permission{
		Name:        "valid-permission-name-" + namePrefix,
		Description: "valid-permission-description-" + namePrefix,
		TimeFields: base.TimeFields{
			CreatedAt: &timeNow,
			UpdatedAt: &timeNow,
		},
	}
	id, createErr := permissionRepoImpl.Create(ctx, permission)
	if createErr != nil {
		t.Errorf("error while creating permission")
	}
	permission.ID = id
	return &permission
}

func TestNewPermissionRepository(t *testing.T) {
	tests := []struct {
		name string
		want PermissionRepository
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPermissionRepository(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPermissionRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionRepositoryImpl_Create(t *testing.T) {
	type args struct {
		ctx        context.Context
		permission entity.Permission
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		args    args
		wantId  int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := tt.repo.Create(tt.args.ctx, tt.args.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("PermissionRepositoryImpl.Create() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func TestPermissionRepositoryImpl_GetByID(t *testing.T) {
	permission := createOnePermission(t, "0043")
	if permission == nil {
		t.Error("failed creating permission : permission is nil")
	}
	defer func() {
		permissionRepoImpl.Delete(ctx, permission.ID)
	}()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name           string
		repo           PermissionRepositoryImpl
		args           args
		wantPermission *entity.Permission
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPermission, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPermission, tt.wantPermission) {
				t.Errorf("PermissionRepositoryImpl.GetByID() = %v, want %v", gotPermission, tt.wantPermission)
			}
		})
	}
}

func TestPermissionRepositoryImpl_GetByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name           string
		repo           PermissionRepositoryImpl
		args           args
		wantPermission *entity.Permission
		wantErr        bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPermission, err := tt.repo.GetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPermission, tt.wantPermission) {
				t.Errorf("PermissionRepositoryImpl.GetByName() = %v, want %v", gotPermission, tt.wantPermission)
			}
		})
	}
}

func TestPermissionRepositoryImpl_GetAll(t *testing.T) {
	type args struct {
		ctx    context.Context
		filter base.RequestGetAll
	}
	tests := []struct {
		name            string
		repo            PermissionRepositoryImpl
		args            args
		wantPermissions []entity.Permission
		wantTotal       int
		wantErr         bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPermissions, gotTotal, err := tt.repo.GetAll(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotPermissions, tt.wantPermissions) {
				t.Errorf("PermissionRepositoryImpl.GetAll() gotPermissions = %v, want %v", gotPermissions, tt.wantPermissions)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("PermissionRepositoryImpl.GetAll() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestPermissionRepositoryImpl_Update(t *testing.T) {
	type args struct {
		ctx        context.Context
		permission entity.Permission
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Update(tt.args.ctx, tt.args.permission); (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPermissionRepositoryImpl_Delete(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		repo    PermissionRepositoryImpl
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("PermissionRepositoryImpl.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
