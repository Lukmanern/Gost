package helper

import (
	"math/rand"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func RandomString(n uint) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func RandomEmails(n uint) []string {
	emailsMap := make(map[string]int)
	for uint(len(emailsMap)) < n {
		body := strings.ToLower(RandomString(7) + RandomString(7) + RandomString(7))
		randEmail := body + "@gost.project"
		emailsMap[randEmail] += 1
	}

	emails := make([]string, 0, len(emailsMap))
	for email := range emailsMap {
		emails = append(emails, email)
	}
	return emails
}

// This used for testing handler : controller/ middleware/ any
func NewFiberCtx() *fiber.Ctx {
	app := fiber.New()
	return app.AcquireCtx(&fasthttp.RequestCtx{})
}
