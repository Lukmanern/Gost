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

func getRbacRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()

	permissionService = service.NewPermissionService()
	permissionController = controller.NewPermissionController(permissionService)
	permissionRouter := router.Group("permission").Use(jwtHandler.IsAuthenticated)

	// create-permission is unused
	permissionRouter.Post("", jwtHandler.CheckHasPermission(rbac.PermCreatePermission.ID), permissionController.Create)
	permissionRouter.Get("", jwtHandler.CheckHasPermission(rbac.PermViewPermission.ID), permissionController.GetAll)
	permissionRouter.Get(":id", jwtHandler.CheckHasPermission(rbac.PermViewPermission.ID), permissionController.Get)
	permissionRouter.Put(":id", jwtHandler.CheckHasPermission(rbac.PermUpdatePermission.ID), permissionController.Update)
	permissionRouter.Delete(":id", jwtHandler.CheckHasPermission(rbac.PermDeletePermission.ID), permissionController.Delete)

	roleService = service.NewRoleService(permissionService)
	roleController = controller.NewRoleController(roleService)
	roleRouter := router.Group("role").Use(jwtHandler.IsAuthenticated)

	roleRouter.Post("", jwtHandler.CheckHasPermission(rbac.PermCreateRole.ID), roleController.Create)
	roleRouter.Post("connect", jwtHandler.CheckHasPermission(rbac.PermCreateRole.ID), roleController.Connect)
	roleRouter.Get("", jwtHandler.CheckHasPermission(rbac.PermViewRole.ID), roleController.GetAll)
	roleRouter.Get(":id", jwtHandler.CheckHasPermission(rbac.PermViewRole.ID), roleController.Get)
	roleRouter.Put(":id", jwtHandler.CheckHasPermission(rbac.PermUpdateRole.ID), roleController.Update)
	roleRouter.Delete(":id", jwtHandler.CheckHasPermission(rbac.PermDeleteRole.ID), roleController.Delete)
}
