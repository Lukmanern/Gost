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

	// create-permission is unused
	permissionRouter.Post("", jwtHandler.CheckHasPermissions(rbac.PermissionCreatePermission), permissionController.Create)
	permissionRouter.Get("", jwtHandler.CheckHasPermissions(rbac.PermissionViewPermission), permissionController.GetAll)
	permissionRouter.Get(":id", jwtHandler.CheckHasPermissions(rbac.PermissionViewPermission), permissionController.Get)
	permissionRouter.Put(":id", jwtHandler.CheckHasPermissions(rbac.PermissionUpdatePermission), permissionController.Update)
	permissionRouter.Delete(":id", jwtHandler.CheckHasPermissions(rbac.PermissionDeletePermission), permissionController.Delete)

	roleService = service.NewRoleService(permissionService)
	roleController = controller.NewRoleController(roleService)
	roleRouter := router.Group("role").Use(jwtHandler.IsAuthenticated)

	roleRouter.Post("", jwtHandler.CheckHasPermissions(rbac.PermissionCreateRole), roleController.Create)
	roleRouter.Post("connect", jwtHandler.CheckHasPermissions(rbac.PermissionCreateRole), roleController.Connect)
	roleRouter.Get("", jwtHandler.CheckHasPermissions(rbac.PermissionViewRole), roleController.GetAll)
	roleRouter.Get(":id", jwtHandler.CheckHasPermissions(rbac.PermissionViewRole), roleController.Get)
	roleRouter.Put(":id", jwtHandler.CheckHasPermissions(rbac.PermissionUpdateRole), roleController.Update)
	roleRouter.Delete(":id", jwtHandler.CheckHasPermissions(rbac.PermissionDeleteRole), roleController.Delete)
}
