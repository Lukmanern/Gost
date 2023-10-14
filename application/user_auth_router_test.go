package application

import (
	"testing"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/gofiber/fiber/v2"
)

func Test_getUserAuthRoutes(t *testing.T) {
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
			name: "Test login route",
			args: args{router: router},
		},
		{
			name: "Test my-profile route",
			args: args{router: router},
		},
		{
			name: "Test logout route",
			args: args{router: router},
		},
		{
			name: "Test update-profile route",
			args: args{router: router},
		},
		{
			name: "Test forget-password route",
			args: args{router: router},
		},
		{
			name: "Test update-password route",
			args: args{router: router},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getUserAuthRoutes(tt.args.router)
			// Logic
		})
	}
}
