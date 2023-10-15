package response

import "github.com/gofiber/fiber/v2"

type response struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// SuccessNoContent formats a successful
// response with HTTP status 204.
func SuccessNoContent(c *fiber.Ctx) error {
	c.Status(fiber.StatusNoContent)
	return c.Send(nil)
}

// CreateResponse generates a new response
// with the given parameters.
func CreateResponse(c *fiber.Ctx, statusCode int, success bool, message string, data interface{}) error {
	c.Status(statusCode)
	return c.JSON(response{
		Message: message,
		Success: success,
		Data:    data,
	})
}

// SuccessLoaded formats a successful response
// with HTTP status 200 and the provided data.
func SuccessLoaded(c *fiber.Ctx, data interface{}) error {
	return CreateResponse(c, fiber.StatusOK, true, "Data successfully loaded", data)
}

// SuccessCreated formats a successful response
// with HTTP status 201 and the provided data.
func SuccessCreated(c *fiber.Ctx, data interface{}) error {
	return CreateResponse(c, fiber.StatusCreated, true, "Data successfully created", data)
}

// BadRequest formats a response with HTTP
// status 400 and the specified message.
func BadRequest(c *fiber.Ctx, message string) error {
	return CreateResponse(c, fiber.StatusBadRequest, false, message, nil)
}

// Unauthorized formats a response with
// HTTP status 401 indicating unauthorized access.
func Unauthorized(c *fiber.Ctx) error {
	return CreateResponse(c, fiber.StatusUnauthorized, false, "Unauthorized", nil)
}

// DataNotFound formats a response with
// HTTP status 404 and the specified message.
func DataNotFound(c *fiber.Ctx, message string) error {
	return CreateResponse(c, fiber.StatusNotFound, false, message, nil)
}

// Error formats an error response
// with HTTP status 500 and the specified message.
func Error(c *fiber.Ctx, message string) error {
	return CreateResponse(c, fiber.StatusInternalServerError, false, message, nil)
}

// ErrorWithData formats an error response
// with HTTP status 500 and the specified
// message and data.
func ErrorWithData(c *fiber.Ctx, message string, data interface{}) error {
	return CreateResponse(c, fiber.StatusInternalServerError, false, message, data)
}
