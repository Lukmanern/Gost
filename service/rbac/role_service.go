package service

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"gorm.io/gorm"
)

type RoleRepository interface {
	Create(ctx context.Context, role entity.Role) (id int, err error)
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error)
	Update(ctx context.Context, role entity.Role) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type RoleRepositoryImpl struct {
	roleTableName string
	db            *gorm.DB
}

var (
	roleTableName          string = "roles"
	roleRepositoryImpl     *RoleRepositoryImpl
	roleRepositoryImplOnce sync.Once
)

func NewRoleRepository() RoleRepository {
	roleRepositoryImplOnce.Do(func() {
		roleRepositoryImpl = &RoleRepositoryImpl{
			roleTableName: roleTableName,
			db:            connector.LoadDatabase(),
		}
	})
	return roleRepositoryImpl
}

func (repo RoleRepositoryImpl) Create(ctx context.Context, role entity.Role) (id int, err error) {
	return
}

func (repo RoleRepositoryImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	return
}

func (repo RoleRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error) {
	return
}

func (repo RoleRepositoryImpl) Update(ctx context.Context, role entity.Role) (err error) {
	return
}

func (repo RoleRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
