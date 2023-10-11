package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/user_auth"
	"github.com/Lukmanern/gost/internal/middleware"
	service "github.com/Lukmanern/gost/service/user_auth"
)

var (
	jwtHandler         *middleware.JWTHandler
	userAuthService    service.UserAuthService
	userAuthController controller.UserAuthController
)

func getUserAuthRoutes(router fiber.Router) {
	jwtHandler = middleware.NewJWTHandler()
	userAuthService = service.NewUserAuthService()
	userAuthController = controller.NewUserAuthController(userAuthService)

	userAuthRoute := router.Group("user/auth")
	userAuthRoute.Post("login", userAuthController.Login)

	userAuthRouteAuth := userAuthRoute.Use(jwtHandler.IsAuthenticated)
	userAuthRouteAuth.Post("logout", userAuthController.Logout)
	userAuthRouteAuth.Get("my-profile", userAuthController.MyProfile)
	userAuthRouteAuth.Post("profile-update", userAuthController.UpdateProfile)
	userAuthRouteAuth.Post("forget-password", userAuthController.ForgetPassword)
	userAuthRouteAuth.Post("update-password", userAuthController.UpdatePassword)
}
