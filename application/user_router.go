// don't use this for production

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user"
	service "github.com/Lukmanern/gost/service/user"
)

var (
	userService    service.UserService
	userController controller.UserController
)

func getUserRoutes(router fiber.Router) {
	userService = service.NewUserService()
	userController = controller.NewUserController(userService)

	userRoute := router.Group("user")
	userRoute.Post("create", userController.Create)
	userRoute.Get("", userController.GetAll)
	userRoute.Get(":id", userController.Get)
	userRoute.Put(":id", userController.Update)
	userRoute.Delete(":id", userController.Delete)
}
