package service

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	repository "github.com/Lukmanern/gost/repository/role"
)

type RoleService interface {
	// Create func create one role.
	Create(ctx context.Context, data model.RoleCreate) (id int, err error)

	// GetByID func get one role.
	GetByID(ctx context.Context, id int) (role model.RoleResponse, err error)

	// GetAll func get some roles.
	GetAll(ctx context.Context, filter model.RequestGetAll) (roles []model.RoleResponse, total int, err error)

	// Update func update one role.
	Update(ctx context.Context, data model.RoleUpdate) (err error)

	// Delete func delete one role.
	Delete(ctx context.Context, id int) (err error)
}

type RoleServiceImpl struct {
	repository repository.RoleRepository
}

var (
	roleServiceImpl     *RoleServiceImpl
	roleServiceImplOnce sync.Once
)

func NewRoleService() RoleService {
	roleServiceImplOnce.Do(func() {
		roleServiceImpl = &RoleServiceImpl{
			repository: repository.NewRoleRepository(),
		}
	})
	return roleServiceImpl
}

func (svc *RoleServiceImpl) Create(ctx context.Context, data model.RoleCreate) (id int, err error) {
	data.Name = strings.ToLower(data.Name)
	role, getErr := svc.repository.GetByName(ctx, data.Name)
	if getErr == nil || role != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "role name has been used")
	}

	entityRole := modelCreateToEntity(data)
	entityRole.SetCreateTime()
	id, err = svc.repository.Create(ctx, entityRole)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (svc *RoleServiceImpl) GetByID(ctx context.Context, id int) (role model.RoleResponse, err error) {
	enttRole, err := svc.repository.GetByID(ctx, id)
	if err == gorm.ErrRecordNotFound {
		return role, fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if err != nil || enttRole == nil {
		return role, errors.New("error while getting role data")
	}
	role = entityToResponse(enttRole)
	return role, nil
}

func (svc *RoleServiceImpl) GetAll(ctx context.Context, filter model.RequestGetAll) (roles []model.RoleResponse, total int, err error) {
	enttRoles, total, err := svc.repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	roles = []model.RoleResponse{}
	for _, enttRole := range enttRoles {
		roles = append(roles, entityToResponse(&enttRole))
	}
	return roles, total, nil
}

func (svc *RoleServiceImpl) Update(ctx context.Context, data model.RoleUpdate) (err error) {
	data.Name = strings.ToLower(data.Name)
	roleByName, getErr := svc.repository.GetByName(ctx, data.Name)
	if getErr != nil && getErr != gorm.ErrRecordNotFound {
		return getErr
	}
	if roleByName != nil && roleByName.ID != data.ID {
		return fiber.NewError(fiber.StatusBadRequest, "role name has been used")
	}

	role, err := svc.repository.GetByID(ctx, data.ID)
	if err == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound)
	}
	if err != nil || role == nil {
		return errors.New("error while getting role data")
	}

	entityRole := modelUpdateToEntity(data)
	entityRole.SetUpdateTime()
	err = svc.repository.Update(ctx, entityRole)
	if err != nil {
		return err
	}
	return nil
}

func (svc *RoleServiceImpl) Delete(ctx context.Context, id int) (err error) {
	role, err := svc.repository.GetByID(ctx, id)
	if err == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound)
	}
	if err != nil || role == nil {
		return errors.New("error while getting role data")
	}

	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func modelCreateToEntity(data model.RoleCreate) entity.Role {
	return entity.Role{
		Name:        data.Name,
		Description: data.Description,
	}
}

func modelUpdateToEntity(data model.RoleUpdate) entity.Role {
	return entity.Role{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
	}
}

func entityToResponse(data *entity.Role) model.RoleResponse {
	return model.RoleResponse{
		ID:          data.ID,
		Name:        data.Name,
		Description: data.Description,
		TimeFields:  data.TimeFields,
	}
}
