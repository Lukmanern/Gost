package service

import (
	"context"
	"sync"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	repository "github.com/Lukmanern/gost/repository/rbac"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PermissionService interface {
	Create(ctx context.Context, permission model.PermissionCreate) (id int, err error)
	GetByID(ctx context.Context, id int) (permission *entity.Permission, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []entity.Permission, total int, err error)
	Update(ctx context.Context, permission model.PermissionUpdate) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type PermissionServiceImpl struct {
	repository repository.PermissionRepository
}

var (
	permissionServiceImpl     *PermissionServiceImpl
	permissionServiceImplOnce sync.Once
)

func NewPermissionService() PermissionService {
	permissionServiceImplOnce.Do(func() {
		permissionServiceImpl = &PermissionServiceImpl{
			repository: repository.NewPermissionRepository(),
		}
	})
	return permissionServiceImpl
}

func (svc PermissionServiceImpl) Create(ctx context.Context, permission model.PermissionCreate) (id int, err error) {
	svc.repository.Create(ctx, entity.Permission{})
	return
}

func (svc PermissionServiceImpl) GetByID(ctx context.Context, id int) (permission *entity.Permission, err error) {
	permission, err = svc.repository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "permission not found")
		}
		return nil, err
	}
	if permission == nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "permission not found")
	}

	return permission, nil
}

func (svc PermissionServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (permissions []entity.Permission, total int, err error) {
	permissions, total, err = svc.repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	if total < 1 {
		return nil, 0, nil
	}

	return permissions, total, nil
}

func (svc PermissionServiceImpl) Update(ctx context.Context, permission model.PermissionUpdate) (err error) {
	svc.repository.Update(ctx, entity.Permission{})
	return
}

func (svc PermissionServiceImpl) Delete(ctx context.Context, id int) (err error) {
	svc.repository.Delete(ctx, id)
	return
}
