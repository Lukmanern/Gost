package application

import (
	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/internal/middleware"

	controller "github.com/Lukmanern/gost/controller/dev"
	service "github.com/Lukmanern/gost/service/email"
)

var (
	emailService  service.EmailService
	devController controller.DevController
)

func getDevRouter(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()
	devController = controller.NewDevControllerImpl()
	// Developement 'helper' Process
	devRouter := router.Group("development")
	devRouter.Get("ping/db", devController.PingDatabase)
	devRouter.Get("ping/redis", devController.PingRedis)
	devRouter.Get("panic", devController.Panic)
	devRouter.Get("new-jwt", devController.NewJWT)
	devRouter.Get("storing-to-redis", devController.StoringToRedis)
	devRouter.Get("get-from-redis", devController.GetFromRedis)

	// userAuthRoute.Use(jwtHandler.IsAuthenticated)
	devAuthRoute := devRouter.Use(jwtHandler.IsAuthenticated)
	devAuthRoute.Get("validate-jwt", devController.ValidateNewJWT)

	// dev email
	emailService = service.NewEmailService()
	emailRoutes := router.Group("email")
	emailRoutes.Post("send-bulk", emailService.Handler)
}
