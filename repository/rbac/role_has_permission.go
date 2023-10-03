package repository

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"gorm.io/gorm"
)

type RoleHasPermissionRepository interface {
	Create(ctx context.Context, user entity.RoleHasPermissions) (id int, err error)
	GetByID(ctx context.Context, id int) (user *entity.RoleHasPermissions, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.RoleHasPermissions, total int, err error)
	Update(ctx context.Context, user entity.RoleHasPermissions) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type RoleHasPermissionRepositoryImpl struct {
	roleHasPermissionTableName string
	db                         *gorm.DB
}

var (
	roleHasPermissionTableName          string = "user_has_roles"
	roleHasPermissionRepositoryImpl     *RoleHasPermissionRepositoryImpl
	roleHasPermissionRepositoryImplOnce sync.Once
)

func NewRoleHasPermissionRepository() RoleHasPermissionRepository {
	roleHasPermissionRepositoryImplOnce.Do(func() {
		roleHasPermissionRepositoryImpl = &RoleHasPermissionRepositoryImpl{
			roleHasPermissionTableName: roleHasPermissionTableName,
			db:                         connector.LoadDatabase(),
		}
	})
	return roleHasPermissionRepositoryImpl
}

func (repo RoleHasPermissionRepositoryImpl) Create(ctx context.Context, user entity.RoleHasPermissions) (id int, err error) {
	return
}

func (repo RoleHasPermissionRepositoryImpl) GetByID(ctx context.Context, id int) (user *entity.RoleHasPermissions, err error) {
	return
}

func (repo RoleHasPermissionRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.RoleHasPermissions, total int, err error) {
	return
}

func (repo RoleHasPermissionRepositoryImpl) Update(ctx context.Context, user entity.RoleHasPermissions) (err error) {
	return
}

func (repo RoleHasPermissionRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
