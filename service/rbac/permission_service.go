package service

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
)

type PermissionService interface {
	Create(ctx context.Context, user entity.Permission) (id int, err error)
	GetByID(ctx context.Context, id int) (user *entity.Permission, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.Permission, total int, err error)
	Update(ctx context.Context, user entity.Permission) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type PermissionServiceImpl struct {
}

var (
	permissionServiceImpl     *PermissionServiceImpl
	permissionServiceImplOnce sync.Once
)

func NewPermissionService() PermissionService {
	permissionServiceImplOnce.Do(func() {
		permissionServiceImpl = &PermissionServiceImpl{}
	})
	return permissionServiceImpl
}

func (service PermissionServiceImpl) Create(ctx context.Context, user entity.Permission) (id int, err error) {
	return
}

func (service PermissionServiceImpl) GetByID(ctx context.Context, id int) (user *entity.Permission, err error) {
	return
}

func (service PermissionServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.Permission, total int, err error) {
	return
}

func (service PermissionServiceImpl) Update(ctx context.Context, user entity.Permission) (err error) {
	return
}

func (service PermissionServiceImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
