package service

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"gorm.io/gorm"
)

type PermissionRepository interface {
	Create(ctx context.Context, user entity.Permission) (id int, err error)
	GetByID(ctx context.Context, id int) (user *entity.Permission, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.Permission, total int, err error)
	Update(ctx context.Context, user entity.Permission) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type PermissionRepositoryImpl struct {
	permissionTableName string
	db                  *gorm.DB
}

var (
	permissionTableName          string = "permissions"
	permissionRepositoryImpl     *PermissionRepositoryImpl
	permissionRepositoryImplOnce sync.Once
)

func NewPermissionRepository() PermissionRepository {
	permissionRepositoryImplOnce.Do(func() {
		permissionRepositoryImpl = &PermissionRepositoryImpl{
			permissionTableName: permissionTableName,
			db:                  connector.LoadDatabase(),
		}
	})
	return permissionRepositoryImpl
}

func (repo PermissionRepositoryImpl) Create(ctx context.Context, user entity.Permission) (id int, err error) {
	return
}

func (repo PermissionRepositoryImpl) GetByID(ctx context.Context, id int) (user *entity.Permission, err error) {
	return
}

func (repo PermissionRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.Permission, total int, err error) {
	return
}

func (repo PermissionRepositoryImpl) Update(ctx context.Context, user entity.Permission) (err error) {
	return
}

func (repo PermissionRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
