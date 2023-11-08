package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"

	controller "github.com/Lukmanern/gost/controller/user"
	service "github.com/Lukmanern/gost/service/user"

	permSvc "github.com/Lukmanern/gost/service/permission"
	roleSvc "github.com/Lukmanern/gost/service/role"
)

var (
	userPermService permSvc.PermissionService
	userRoleService roleSvc.RoleService
	userService     service.UserService
	userController  controller.UserController
)

func getUserRoutes(router fiber.Router) {
	userPermService = permSvc.NewPermissionService()
	userRoleService = roleSvc.NewRoleService(userPermService)
	userService = service.NewUserService(userRoleService)
	userController = controller.NewUserController(userService)
	jwtHandler := middleware.NewJWTHandler()

	userRoute := router.Group("user")
	userRoute.Post("login", userController.Login)
	userRoute.Post("register", userController.Register)
	userRoute.Post("verification", userController.AccountActivation)
	userRoute.Post("request-delete", userController.DeleteAccountActivation)
	userRoute.Post("forget-password", userController.ForgetPassword)
	userRoute.Post("reset-password", userController.ResetPassword)

	userRouteAuth := userRoute.Use(jwtHandler.IsAuthenticated)
	userRouteAuth.Post("logout", userController.Logout)
	userRouteAuth.Get("my-profile", userController.MyProfile)
	userRouteAuth.Put("profile-update", userController.UpdateProfile)
	userRouteAuth.Post("update-password", userController.UpdatePassword)
}
