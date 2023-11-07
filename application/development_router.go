// don't use this for production
// use this file just for testing
// and testing management.

package application

import (
	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/development"
	"github.com/Lukmanern/gost/internal/middleware"
)

var (
	devController controller.DevController
)

func getDevopmentRouter(router fiber.Router) {
	jwtHandler := middleware.NewJWTHandler()
	devController = controller.NewDevControllerImpl()
	// Developement 'helper' Process
	devRouter := router.Group("development")
	devRouter.Get("ping/db", devController.PingDatabase)
	devRouter.Get("ping/redis", devController.PingRedis)
	devRouter.Get("panic", devController.Panic)
	devRouter.Get("storing-to-redis", devController.StoringToRedis)
	devRouter.Get("get-from-redis", devController.GetFromRedis)
	devRouter.Post("upload-file", devController.UploadFile)
	devRouter.Post("get-files-list", devController.GetFilesList)

	// you should create new role named new-role-001 and new permission
	// named new-permission-001 from RBAC-endpoints to test these endpoints
	devRouterAuth := devRouter.Use(jwtHandler.IsAuthenticated)
	devRouterAuth.Get("test-new-role",
		jwtHandler.CheckHasRole("new-role-001"), devController.CheckNewRole)
	devRouterAuth.Get("test-new-permission",
		jwtHandler.CheckHasPermission(21), devController.CheckNewPermission)
}
