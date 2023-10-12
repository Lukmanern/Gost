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

type RoleService interface {
	Create(ctx context.Context, data model.RoleCreate) (id int, err error)
	ConnectPermissions(ctx context.Context, data model.RoleConnectToPermissions) (err error)
	GetByID(ctx context.Context, id int) (role *entity.Role, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error)
	Update(ctx context.Context, data model.RoleUpdate) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type RoleServiceImpl struct {
	repository        repository.RoleRepository
	servicePermission PermissionService
}

var (
	roleServiceImpl     *RoleServiceImpl
	roleServiceImplOnce sync.Once
)

func NewRoleService(servicePermission PermissionService) RoleService {
	roleServiceImplOnce.Do(func() {
		roleServiceImpl = &RoleServiceImpl{
			repository:        repository.NewRoleRepository(),
			servicePermission: servicePermission,
		}
	})
	return roleServiceImpl
}

func (svc RoleServiceImpl) Create(ctx context.Context, data model.RoleCreate) (id int, err error) {
	// roleEntity
	// svc.repository.GetAll(ctx, base.RequestGetAll{Limit: })
	return
}

func (svc RoleServiceImpl) ConnectPermissions(ctx context.Context, data model.RoleConnectToPermissions) (err error) {
	role, getErr := svc.repository.GetByID(ctx, data.RoleID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}
		return getErr
	}
	if role == nil {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}
	for _, id := range data.PermissionsID {
		service, getErr := svc.servicePermission.GetByID(ctx, id)
		if getErr != nil || service == nil {
			return fiber.NewError(fiber.StatusNotFound, "on of services isn't found")
		}
	}

	deleteErr := svc.DeleteRoleHasPermissions(ctx, data.RoleID)
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}

func (svc RoleServiceImpl) GetByID(ctx context.Context, id int) (role *entity.Role, err error) {
	role, err = svc.repository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "role not found")
		}
		return nil, err
	}
	if role == nil {
		return nil, fiber.NewError(fiber.StatusNotFound, "role not found")
	}

	return role, nil
}

func (svc RoleServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (roles []entity.Role, total int, err error) {
	return
}

func (svc RoleServiceImpl) Update(ctx context.Context, data model.RoleUpdate) (err error) {
	role, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}
		return getErr
	}
	if role == nil {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}

	entityRole := entity.Role{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
	}
	entityRole.SetUpdateTime()
	err = svc.repository.Update(ctx, entityRole)
	if err != nil {
		return err
	}
	return nil
}

func (svc RoleServiceImpl) Delete(ctx context.Context, id int) (err error) {
	role, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}
		return getErr
	}
	if role == nil {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}
	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (svc RoleServiceImpl) DeleteRoleHasPermissions(ctx context.Context, id int) (err error) {
	role, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "role not found")
		}
		return getErr
	}
	if role == nil {
		return fiber.NewError(fiber.StatusNotFound, "role not found")
	}
	err = svc.repository.DeleteRoleHasPermissions(ctx, id)
	if err != nil {
		return err
	}
	return nil
}
