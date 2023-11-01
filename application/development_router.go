// don't use this for production
// use this file just for testing
// and testing management.

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/dev"
)

var (
	devController controller.DevController
)

func getDevopmentRouter(router fiber.Router) {
	devController = controller.NewDevControllerImpl()
	// Developement 'helper' Process
	devRouter := router.Group("development")
	devRouter.Get("ping/db", devController.PingDatabase)
	devRouter.Get("ping/redis", devController.PingRedis)
	devRouter.Get("panic", devController.Panic)
	devRouter.Get("storing-to-redis", devController.StoringToRedis)
	devRouter.Get("get-from-redis", devController.GetFromRedis)
}
