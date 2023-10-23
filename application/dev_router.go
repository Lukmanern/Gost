package application

import (
	controller "github.com/Lukmanern/gost/controller/dev"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

var devController controller.DevController

func getDevRouter(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()
	devController = controller.NewDevControllerImpl()
	// Developement 'helper' Process
	devRouter := router.Group("development")
	devRouter.Get("bitfield", devController.BitfieldTesting)
	devRouter.Get("ping/mysql", devController.PingDatabase)
	devRouter.Get("ping/redis", devController.PingRedis)
	devRouter.Get("panic", devController.Panic)
	devRouter.Get("new-jwt", devController.NewJWT)

	// userAuthRoute.Use(jwtHandler.IsAuthenticated)
	devAuthRoute := devRouter.Use(jwtHandler.IsAuthenticated)
	devAuthRoute.Get("validate-jwt", devController.ValidateNewJWT)
}
