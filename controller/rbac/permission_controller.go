package service

import (
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/base"
	service "github.com/Lukmanern/gost/service/rbac"
)

type PermissionController interface {
	Create(c *fiber.Ctx) error
	Get(c *fiber.Ctx) error
	GetAll(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
}

type PermissionControllerImpl struct {
	service service.PermissionService
}

var (
	permissionControllerImpl     *PermissionControllerImpl
	permissionControllerImplOnce sync.Once
)

func NewPermissionController(service service.PermissionService) PermissionController {
	permissionControllerImplOnce.Do(func() {
		permissionControllerImpl = &PermissionControllerImpl{
			service: service,
		}
	})
	return permissionControllerImpl
}

func (ctr PermissionControllerImpl) Create(c *fiber.Ctx) error {
	return base.ResponseCreated(c, "success created", nil)
}

func (ctr PermissionControllerImpl) Get(c *fiber.Ctx) error {
	return base.ResponseLoaded(c, nil)
}

func (ctr PermissionControllerImpl) GetAll(c *fiber.Ctx) error {
	return base.ResponseLoaded(c, nil)
}

func (ctr PermissionControllerImpl) Update(c *fiber.Ctx) error {
	return base.ResponseNoContent(c)
}

func (ctr PermissionControllerImpl) Delete(c *fiber.Ctx) error {
	return base.ResponseNoContent(c)
}
