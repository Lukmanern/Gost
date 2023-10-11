package application

import (
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/rbac"
	service "github.com/Lukmanern/gost/service/rbac"
)

var (
	permissionService    service.PermissionService
	permissionController controller.PermissionController

	roleService    service.RoleService
	roleController controller.RoleController
)

func getRBACAuthRoutes(router fiber.Router) {
	_ = middleware.NewJWTHandler()
}
