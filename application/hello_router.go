package application

import (
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
)

func helloRoutes(router fiber.Router) {
	helloRouter := router.Group("hello")
	helloRouter.Get("", func(c *fiber.Ctx) error {
		return response.SuccessLoaded(c, "Hello from Backend !")
	})
	helloRouter.Get(":name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		return response.SuccessLoaded(c, "Hello "+name)
	})
}
