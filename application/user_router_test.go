package application

import (
	"testing"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/gofiber/fiber/v2"
)

func Test_getUserRoutes(t *testing.T) {
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
			name: "Test create user route",
			args: args{router: router},
		},
		{
			name: "Test get all users route",
			args: args{router: router},
		},
		{
			name: "Test get user by ID route",
			args: args{router: router},
		},
		{
			name: "Test update user route",
			args: args{router: router},
		},
		{
			name: "Test delete user route",
			args: args{router: router},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getUserRoutes(tt.args.router)
			// Logic
		})
	}
}
