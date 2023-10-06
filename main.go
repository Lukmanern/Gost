package main

import (
	"github.com/Lukmanern/gost/application"
	"github.com/Lukmanern/gost/internal/env"
)

func main() {
	env.ReadConfig("./.env")
	_ = env.Configuration()

	// exp := time.Now().Add(5 * time.Hour)
	// fmt.Println(jwt.NewJWTHandler().GenerateJWT(1, "email", "role", []string{"any-1", "any-2"}, exp))

	application.RunApp()
}
