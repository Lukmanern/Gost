package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"

	controller "github.com/Lukmanern/gost/controller/rbac"
	service "github.com/Lukmanern/gost/service/rbac"
)

var (
	roleService    service.RoleService
	roleController controller.RoleController

	permissionService    service.PermissionService
	permissionController controller.PermissionController
)

func getRBACAuthRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()

	permissionService = service.NewPermissionService()
	permissionController = controller.NewPermissionController(permissionService)
	permissionRouter := router.Group("permission").Use(jwtHandler.IsAuthenticated)

	permissionRouter.Post("", jwtHandler.CheckHasPermission(rbac.PermissionCreatePermission), permissionController.Create)
	permissionRouter.Get("", jwtHandler.CheckHasPermission(rbac.PermissionViewPermission), permissionController.GetAll)
	permissionRouter.Get(":id", jwtHandler.CheckHasPermission(rbac.PermissionViewPermission), permissionController.Get)
	permissionRouter.Put(":id", jwtHandler.CheckHasPermission(rbac.PermissionUpdatePermission), permissionController.Update)
	permissionRouter.Delete(":id", jwtHandler.CheckHasPermission(rbac.PermissionDeletePermission), permissionController.Delete)

	roleService = service.NewRoleService(permissionService)
	roleController = controller.NewRoleController(roleService)
	roleRouter := router.Group("role").Use(jwtHandler.IsAuthenticated)

	roleRouter.Post("", jwtHandler.CheckHasPermission(rbac.PermissionCreateRole), roleController.Create)
	roleRouter.Post("connect", jwtHandler.CheckHasPermission(rbac.PermissionCreateRole), roleController.Connect)
	roleRouter.Get("", jwtHandler.CheckHasPermission(rbac.PermissionViewRole), roleController.GetAll)
	roleRouter.Get(":id", jwtHandler.CheckHasPermission(rbac.PermissionViewRole), roleController.Get)
	roleRouter.Put(":id", jwtHandler.CheckHasPermission(rbac.PermissionUpdateRole), roleController.Update)
	roleRouter.Delete(":id", jwtHandler.CheckHasPermission(rbac.PermissionDeleteRole), roleController.Delete)
}
