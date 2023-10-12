package service

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"gorm.io/gorm"
)

type RoleService interface {
	Create(ctx context.Context, data model.RoleCreate) (id int, err error)
	ConnectPermissions(ctx context.Context, data model.RoleConnectToPermissions) (err error)
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error)
	Update(ctx context.Context, data model.RoleUpdate) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type RoleServiceImpl struct {
	roleTableName string
	db            *gorm.DB
}

var (
	roleTableName       string = "roles"
	roleServiceImpl     *RoleServiceImpl
	roleServiceImplOnce sync.Once
)

func NewRoleService() RoleService {
	roleServiceImplOnce.Do(func() {
		roleServiceImpl = &RoleServiceImpl{
			roleTableName: roleTableName,
			db:            connector.LoadDatabase(),
		}
	})
	return roleServiceImpl
}

func (repo RoleServiceImpl) Create(ctx context.Context, data model.RoleCreate) (id int, err error) {
	return
}

func (repo RoleServiceImpl) ConnectPermissions(ctx context.Context, data model.RoleConnectToPermissions) (err error) {
	// search role
	// loop : search all permissions
	return nil
}

func (repo RoleServiceImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	return
}

func (repo RoleServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error) {
	return
}

func (repo RoleServiceImpl) Update(ctx context.Context, data model.RoleUpdate) (err error) {
	return
}

func (repo RoleServiceImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
