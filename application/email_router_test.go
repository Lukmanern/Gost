package application

import (
	"testing"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/gofiber/fiber/v2"
)

func Test_getEmailRouter(t *testing.T) {
	env.ReadConfig("./../.env")
	router := fiber.New()

	type args struct {
		router fiber.Router
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test send-bulk email route",
			args: args{router: router},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getEmailRouter(tt.args.router)
			// Logic
		})
	}
}
