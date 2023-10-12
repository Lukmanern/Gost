package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"

	controller "github.com/Lukmanern/gost/controller/user_auth"
	service "github.com/Lukmanern/gost/service/user_auth"
)

var (
	userAuthService    service.UserAuthService
	userAuthController controller.UserAuthController
)

func getUserAuthRoutes(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()
	userAuthService = service.NewUserAuthService()
	userAuthController = controller.NewUserAuthController(userAuthService)

	userAuthRoute := router.Group("user/auth")
	userAuthRoute.Post("login", userAuthController.Login)

	userAuthRouteAuth := userAuthRoute.Use(jwtHandler.IsAuthenticated)
	userAuthRouteAuth.Get("my-profile", userAuthController.MyProfile)
	userAuthRouteAuth.Post("logout", userAuthController.Logout)
	userAuthRouteAuth.Post("profile-update", userAuthController.UpdateProfile)
	userAuthRouteAuth.Post("forget-password", userAuthController.ForgetPassword)
	userAuthRouteAuth.Post("update-password", userAuthController.UpdatePassword)
}
