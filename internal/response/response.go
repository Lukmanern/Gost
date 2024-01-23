package response

import (
	"strings"

	"github.com/Lukmanern/gost/internal/consts"
	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// CreateResponse generates a new response
// with the given parameters.
func CreateResponse(c *fiber.Ctx, statusCode int, response Response) error {
	c.Status(statusCode)
	return c.JSON(Response{
		Message: strings.ToLower(response.Message),
		Success: response.Success,
		Data:    response.Data,
	})
}

// SuccessNoContent formats a successful
// response with HTTP status 204.
func SuccessNoContent(c *fiber.Ctx) error {
	c.Status(fiber.StatusNoContent)
	return c.Send(nil)
}

// SuccessLoaded formats a successful response
// with HTTP status 200 and the provided data.
func SuccessLoaded(c *fiber.Ctx, data interface{}) error {
	return CreateResponse(c, fiber.StatusOK, Response{
		Message: strings.ToLower(consts.SuccessLoaded),
		Success: true,
		Data:    data,
	})
}

// SuccessCreated formats a successful response
// with HTTP status 201 and the provided data.
func SuccessCreated(c *fiber.Ctx, data interface{}) error {
	return CreateResponse(c, fiber.StatusCreated, Response{
		Message: strings.ToLower(consts.SuccessCreated),
		Success: true,
		Data:    data,
	})
}

// BadRequest formats a response with HTTP status 400.
func BadRequest(c *fiber.Ctx) error {
	return CreateResponse(c, fiber.StatusBadRequest, Response{
		Message: consts.BadRequest,
		Success: false,
		Data:    nil,
	})
}

// Unauthorized formats a response with
// HTTP status 401 indicating unauthorized access.
func Unauthorized(c *fiber.Ctx) error {
	return CreateResponse(c, fiber.StatusUnauthorized, Response{
		Message: consts.Unauthorized,
		Success: false,
		Data:    nil,
	})
}

// DataNotFound formats a response with
// HTTP status 404 and the specified message.
func DataNotFound(c *fiber.Ctx) error {
	return CreateResponse(c, fiber.StatusNotFound, Response{
		Message: consts.NotFound,
		Success: false,
		Data:    nil,
	})
}

// Error formats an error response
// with HTTP status 500 and the specified message.
func Error(c *fiber.Ctx, message string) error {
	return CreateResponse(c, fiber.StatusInternalServerError, Response{
		Message: message,
		Success: false,
		Data:    nil,
	})
}

// ErrorWithData formats an error response
// with HTTP status 500 and the specified
// message and data.
func ErrorWithData(c *fiber.Ctx, message string, data interface{}) error {
	return CreateResponse(c, fiber.StatusInternalServerError, Response{
		Message: message,
		Success: false,
		Data:    data,
	})
}
