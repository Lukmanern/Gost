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
	roleRepoImpl RoleRepositoryImpl
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)
	timeNow = time.Now()
	ctx = context.Background()
	roleRepoImpl = RoleRepositoryImpl{
		roleTableName: roleTableName,
		db:            connector.LoadDatabase(),
	}

}

func TestNewRoleRepository(t *testing.T) {
	tests := []struct {
		name string
		want RoleRepository
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRoleRepository(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRoleRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleRepositoryImpl_Create(t *testing.T) {
	type args struct {
		ctx           context.Context
		role          entity.Role
		permissionsID []int
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantId  int
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotId, err := tt.repo.Create(tt.args.ctx, tt.args.role, tt.args.permissionsID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId != tt.wantId {
				t.Errorf("RoleRepositoryImpl.Create() = %v, want %v", gotId, tt.wantId)
			}
		})
	}
}

func TestRoleRepositoryImpl_ConnectToPermission(t *testing.T) {
	type args struct {
		ctx           context.Context
		roleID        int
		permissionsID []int
	}
	tests := []struct {
		name    string
		repo    RoleRepositoryImpl
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.ConnectToPermission(tt.args.ctx, tt.args.roleID, tt.args.permissionsID); (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.ConnectToPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoleRepositoryImpl_GetByID(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name     string
		repo     RoleRepositoryImpl
		args     args
		wantRole *entity.Role
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRole, err := tt.repo.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRole, tt.wantRole) {
				t.Errorf("RoleRepositoryImpl.GetByID() = %v, want %v", gotRole, tt.wantRole)
			}
		})
	}
}

func TestRoleRepositoryImpl_GetByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name     string
		repo     RoleRepositoryImpl
		args     args
		wantRole *entity.Role
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRole, err := tt.repo.GetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRole, tt.wantRole) {
				t.Errorf("RoleRepositoryImpl.GetByName() = %v, want %v", gotRole, tt.wantRole)
			}
		})
	}
}

func TestRoleRepositoryImpl_GetAll(t *testing.T) {
	type args struct {
		ctx    context.Context
		filter base.RequestGetAll
	}
	tests := []struct {
		name      string
		repo      RoleRepositoryImpl
		args      args
		wantRoles []entity.Role
		wantTotal int
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRoles, gotTotal, err := tt.repo.GetAll(tt.args.ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRoles, tt.wantRoles) {
				t.Errorf("RoleRepositoryImpl.GetAll() gotRoles = %v, want %v", gotRoles, tt.wantRoles)
			}
			if gotTotal != tt.wantTotal {
				t.Errorf("RoleRepositoryImpl.GetAll() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
		})
	}
}

func TestRoleRepositoryImpl_Update(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Update(tt.args.ctx, tt.args.role); (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoleRepositoryImpl_Delete(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RoleRepositoryImpl.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
