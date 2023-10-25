// don't use this for production

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user_dev"
	service "github.com/Lukmanern/gost/service/user_dev"
)

var (
	userService    service.UserDevService
	userController controller.UserController
)

func getUserRoutes(router fiber.Router) {
	userService = service.NewUserDevService()
	userController = controller.NewUserController(userService)

	userRoute := router.Group("user/dev")
	userRoute.Post("create", userController.Create)
	userRoute.Get("", userController.GetAll)
	userRoute.Get(":id", userController.Get)
	userRoute.Put(":id", userController.Update)
	userRoute.Delete(":id", userController.Delete)
}
