package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user_auth"
	service "github.com/Lukmanern/gost/service/user_auth"
)

var (
	userAuthService    service.UserAuthService
	userAuthController controller.UserAuthController
)

func getUserAuthRoutes(router fiber.Router) {
	userAuthService = service.NewUserAuthService()
	userAuthController = controller.NewUserAuthController(userAuthService)

	userAuthRoute := router.Group("user/auth")
	// login-logout
	userAuthRoute.Post("login", userAuthController.Login)
	userAuthRoute.Post("logout", userAuthController.Logout)
	// profile
	userAuthRoute.Post("profile-update", userAuthController.UpdateProfile)
	// update password
	userAuthRoute.Post("forget-password", userAuthController.ForgetPassword)
	userAuthRoute.Post("update-password", userAuthController.UpdatePassword)
}
