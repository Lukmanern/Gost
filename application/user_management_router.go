// don't use this for production
// use this file just for testing
// and testing management.

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user_management"
	service "github.com/Lukmanern/gost/service/user_management"
)

var (
	userDevService    service.UserManagementService
	userDevController controller.UserManagementController
)

func getUserManagementRoutes(router fiber.Router) {
	userDevService = service.NewUserManagementService()
	userDevController = controller.NewUserManagementController(userDevService)

	userDevRoute := router.Group("user-management")
	userDevRoute.Post("create", userDevController.Create)
	userDevRoute.Get("", userDevController.GetAll)
	userDevRoute.Get(":id", userDevController.Get)
	userDevRoute.Put(":id", userDevController.Update)
	userDevRoute.Delete(":id", userDevController.Delete)
}
