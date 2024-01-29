// 📌 Origin Github Repository: https://github.com/Lukmanern

package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/role"

	roleCtr "github.com/Lukmanern/gost/controller/role"
	roleSvc "github.com/Lukmanern/gost/service/role"
)

var (
	roleService    roleSvc.RoleService
	roleController roleCtr.RoleController
)

func getRolePermissionRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()

	roleService = roleSvc.NewRoleService()
	roleController = roleCtr.NewRoleController(roleService)

	roleRouter := router.Group("role").Use(jwtHandler.IsAuthenticated)
	roleRouter.Post("", jwtHandler.HasRole(role.RoleSuperAdmin), roleController.Create)
	roleRouter.Get("", jwtHandler.HasRole(role.RoleSuperAdmin), roleController.GetAll)
	roleRouter.Get(":id", jwtHandler.HasRole(role.RoleSuperAdmin), roleController.Get)
	roleRouter.Put(":id", jwtHandler.HasRole(role.RoleSuperAdmin), roleController.Update)
	roleRouter.Delete(":id", jwtHandler.HasRole(role.RoleSuperAdmin), roleController.Delete)
}
