// 📌 Origin Github Repository: https://github.com/Lukmanern

package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/role"

	controller "github.com/Lukmanern/gost/controller/role"
	service "github.com/Lukmanern/gost/service/role"
)

var (
	roleService    service.RoleService
	roleController controller.RoleController
)

func roleRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()

	roleService = service.NewRoleService()
	roleController = controller.NewRoleController(roleService)

	roleRouter := router.Group("role").Use(jwtHandler.IsAuthenticated)
	roleRouter.Post("", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), roleController.Create)
	roleRouter.Get("", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), roleController.GetAll)
	roleRouter.Get(":id", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), roleController.Get)
	roleRouter.Put(":id", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), roleController.Update)
	roleRouter.Delete(":id", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), roleController.Delete)
}
