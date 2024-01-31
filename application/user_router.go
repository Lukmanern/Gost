// 📌 Origin Github Repository: https://github.com/Lukmanern

package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/role"

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
	userRoute.Post("register", userController.Register) // send email
	userRoute.Post("account-activation", userController.AccountActivation)
	userRoute.Post("login", userController.Login)
	userRoute.Post("forget-password", userController.ForgetPassword) // send email
	userRoute.Post("reset-password", userController.ResetPassword)

	userRouteAuth := userRoute.Use(jwtHandler.IsAuthenticated)
	userRouteAuth.Get("my-profile", userController.MyProfile)
	userRouteAuth.Post("logout", userController.Logout)
	userRouteAuth.Put("profile-update", userController.UpdateProfile)
	userRouteAuth.Post("update-password", userController.UpdatePassword)
	userRouteAuth.Delete("delete-account", userController.DeleteAccount)

	// for admin
	userRouteAuth.Get("", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), userController.GetAll)
	userRouteAuth.Put("ban-user/:id", jwtHandler.HasOneRole(role.RoleSuperAdmin, role.RoleAdmin), userController.BanAccount)
}
