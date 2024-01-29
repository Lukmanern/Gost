// 📌 Origin Github Repository: https://github.com/Lukmanern

package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"

	controller "github.com/Lukmanern/gost/controller/user"
	service "github.com/Lukmanern/gost/service/user"
)

var (
	userService    service.UserService
	userController controller.UserController
)

func getUserRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()

	userService = service.NewUserService()
	userController = controller.NewUserController(userService)

	userRoute := router.Group("user")
	userRoute.Post("login", userController.Login)
	userRoute.Post("register", userController.Register)
	userRoute.Post("verification", userController.AccountActivation)
	userRoute.Post("forget-password", userController.ForgetPassword)
	userRoute.Post("reset-password", userController.ResetPassword)

	// get all
	userRouteAuth := userRoute.Use(jwtHandler.IsAuthenticated)
	userRouteAuth.Post("logout", userController.Logout)
	userRouteAuth.Get("my-profile", userController.MyProfile)
	userRouteAuth.Put("profile-update", userController.UpdateProfile)
	userRouteAuth.Post("update-password", userController.UpdatePassword)
}
