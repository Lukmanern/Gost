package application

import (
	controller "github.com/Lukmanern/gost/controller/dev"
	"github.com/gofiber/fiber/v2"
)

var devController controller.DevController

func getDevRouter(router fiber.Router) {
	devController = controller.NewDevControllerImpl()
	// Developement 'helper' Process
	devRouter := router.Group("development")
	devRouter.Get("bitfield", devController.BitfieldTesting)
	devRouter.Get("ping/mysql", devController.PingDatabase)
	devRouter.Get("ping/redis", devController.PingRedis)
	devRouter.Get("panic", devController.Panic)
}
