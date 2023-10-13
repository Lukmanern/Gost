package base

import (
	"github.com/gofiber/fiber/v2"
)

type PageMeta struct {
	Total int `json:"total"`
	Pages int `json:"pages"`
	Page  int `json:"page"`
}

type GetAllResponse struct {
	Meta PageMeta      `json:"meta"`
	Data []interface{} `json:"data"`
}

type response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func formatResponse(c *fiber.Ctx, statusCode int, success bool, message string, data interface{}) error {
	c.Status(statusCode)
	return c.JSON(response{
		Message: message,
		Success: success,
		Data:    data,
	})
}

// FormatResponse is a generic response formatter.
func Response(c *fiber.Ctx, statusCode int, success bool, message string, data interface{}) error {
	return formatResponse(c, statusCode, success, message, data)
}

// FormatResponseLoaded formats a successful
// response with HTTP status 200.
func ResponseLoaded(c *fiber.Ctx, data interface{}) error {
	return formatResponse(
		c, fiber.StatusOK, true, "success get data", data)
}

// FormatResponseLoaded formats a successful
// response with HTTP status 200.
func ResponseUpdated(c *fiber.Ctx, message string) error {
	if message == "" {
		message = "data success updated"
	}
	return formatResponse(
		c, fiber.StatusOK, true, message, nil)
}

// FormatResponseCreated formats a successful
// response with HTTP status 201.
func ResponseCreated(c *fiber.Ctx, message string, data interface{}) error {
	return formatResponse(
		c, fiber.StatusCreated, true, message, data)
}

// ResponseNoContent formats a successful
// response with HTTP status 204.
func ResponseNoContent(c *fiber.Ctx) error {
	c.Status(fiber.StatusNoContent)
	return c.Send(nil)
}

// FormatResponseBadRequest formats a
// response with HTTP status 400.
func ResponseBadRequest(c *fiber.Ctx, message string) error {
	return formatResponse(
		c, fiber.StatusBadRequest, false, message, nil)
}

// FormatResponseUnauthorized formats a
// response with HTTP status 401.
func ResponseUnauthorized(c *fiber.Ctx) error {
	return formatResponse(
		c, fiber.StatusUnauthorized, false, "unauthorized", nil)
}

// FormatResponseNotFound formats a
// response with HTTP status 404.
func ResponseNotFound(c *fiber.Ctx, message string) error {
	return formatResponse(
		c, fiber.StatusNotFound, false, message, nil)
}

// FormatResponseInternalServerError formats a
// response with HTTP status 500.
func ResponseInternalServerError(c *fiber.Ctx, message string) error {
	return formatResponse(
		c, fiber.StatusInternalServerError, false, message, nil)
}

func ResponseErrorWithData(c *fiber.Ctx, message string, data interface{}) error {
	return formatResponse(
		c, fiber.StatusInternalServerError, false, message, data)
}
