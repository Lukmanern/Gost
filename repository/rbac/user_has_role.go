package repository

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"gorm.io/gorm"
)

type UserHasRoleRepository interface {
	Create(ctx context.Context, user entity.UserHasRoles) (id int, err error)
	GetByID(ctx context.Context, id int) (user *entity.UserHasRoles, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.UserHasRoles, total int, err error)
	Update(ctx context.Context, user entity.UserHasRoles) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type UserHasRoleRepositoryImpl struct {
	userHasRoleTableName string
	db                   *gorm.DB
}

var (
	userHasRoleTableName          string = "user_has_roles"
	userHasRoleRepositoryImpl     *UserHasRoleRepositoryImpl
	userHasRoleRepositoryImplOnce sync.Once
)

func NewUserHasRoleRepository() UserHasRoleRepository {
	userHasRoleRepositoryImplOnce.Do(func() {
		userHasRoleRepositoryImpl = &UserHasRoleRepositoryImpl{
			userHasRoleTableName: userHasRoleTableName,
			db:                   connector.LoadDatabase(),
		}
	})
	return userHasRoleRepositoryImpl
}

func (repo UserHasRoleRepositoryImpl) Create(ctx context.Context, user entity.UserHasRoles) (id int, err error) {
	return
}

func (repo UserHasRoleRepositoryImpl) GetByID(ctx context.Context, id int) (user *entity.UserHasRoles, err error) {
	return
}

func (repo UserHasRoleRepositoryImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []entity.UserHasRoles, total int, err error) {
	return
}

func (repo UserHasRoleRepositoryImpl) Update(ctx context.Context, user entity.UserHasRoles) (err error) {
	return
}

func (repo UserHasRoleRepositoryImpl) Delete(ctx context.Context, id int) (err error) {
	return
}
