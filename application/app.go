package application

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/Lukmanern/gost/database/connector"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var (
	// Create a new fiber instance with custom config
	router = fiber.New(fiber.Config{
		// Override default error handler
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			// Send custom error page
			err = ctx.Status(code).JSON(fiber.Map{
				"message": e.Message,
			})
			if err != nil {
				// In case the SendFile fails
				return ctx.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
			}

			// Return from handler
			return nil
		},
		ReadBufferSize: 12000,
	})
)

func RunApp() {
	connector.LoadDatabase()

	router.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	router.Use(logger.New())
	// Custom File Writer
	_ = os.MkdirAll("./log", os.ModePerm)
	file, err := os.OpenFile(fmt.Sprintf("./log/%s.log", time.Now().Format("20060102")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()
	router.Use(logger.New(logger.Config{
		Output: file,
	}))

	// Create channel for idle connections.
	idleConnsClosed := make(chan struct{})

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt) // Catch OS signals.
		<-sigint

		// Received an interrupt signal, shutdown.
		if err := router.Shutdown(); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("Oops... Server is not shutting down! Reason: %v", err)
		}

		close(idleConnsClosed)
	}()

	getUserRoutes(router)
	getUserAuthRoutes(router)

	if err := router.Listen(fmt.Sprintf(":%d", 9009)); err != nil {
		log.Printf("Oops... Server is not running! Reason: %v", err)
	}

	<-idleConnsClosed
}
