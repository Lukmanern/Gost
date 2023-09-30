package base

import "github.com/gofiber/fiber/v2"

type StdFormatReturn struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// this func used by controller-layer
func FormatResponse(c *fiber.Ctx, statusCode int, success bool, data interface{}, message string) error {
	c.Status(statusCode)
	return c.JSON(StdFormatReturn{
		Message: message,
		Success: success,
		Data:    data,
	})
}
