// don't use this for production
// use this file just for testing
// and testing management.

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user_dev"
	service "github.com/Lukmanern/gost/service/user_dev"
)

var (
	userDevService    service.UserDevService
	userDevController controller.UserController
)

func getUserDevRoutes(router fiber.Router) {
	userDevService = service.NewUserDevService()
	userDevController = controller.NewUserController(userDevService)

	userDevRoute := router.Group("user-management")
	userDevRoute.Post("create", userDevController.Create)
	userDevRoute.Get("", userDevController.GetAll)
	userDevRoute.Get(":id", userDevController.Get)
	userDevRoute.Put(":id", userDevController.Update)
	userDevRoute.Delete(":id", userDevController.Delete)
}
