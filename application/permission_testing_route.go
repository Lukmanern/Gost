package application

import (
	"github.com/Lukmanern/gost/internal/response"
	"github.com/gofiber/fiber/v2"
)

func getPermissionTestingRoute(router fiber.Router) {

}

func FakeHandler(c *fiber.Ctx) error {
	return response.SuccessLoaded(c, nil)
}
