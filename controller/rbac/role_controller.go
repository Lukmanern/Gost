package service

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
)

type RoleController interface {
	Create(ctx context.Context, role entity.Role) (id int, err error)
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error)
	Update(ctx context.Context, role entity.Role) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type RoleControllerImpl struct {
}

var (
	roleControllerImpl     *RoleControllerImpl
	roleControllerImplOnce sync.Once
)

func NewRoleController() RoleController {
	roleControllerImplOnce.Do(func() {
		roleControllerImpl = &RoleControllerImpl{}
	})
	return roleControllerImpl
}

func (repo RoleControllerImpl) Create(ctx context.Context, role entity.Role) (id int, err error) {
	return
}

func (repo RoleControllerImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	return
}

func (repo RoleControllerImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error) {
	return
}

func (repo RoleControllerImpl) Update(ctx context.Context, role entity.Role) (err error) {
	return
}

func (repo RoleControllerImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
