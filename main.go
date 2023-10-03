package main

import (
	"fmt"
	"time"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/jwt"
)

func main() {
	env.ReadConfig("./.env")
	config := env.Configuration()
	config.ShowConfig()

	expired := time.Now().Add(config.AppAccessTokenTTL)
	jwtHandler := jwt.NewHWTHandler()

	fmt.Println(jwtHandler.GenerateJWT(1, "xxx", "xxxx", []string{"xxx"}, expired))
	fmt.Println(jwtHandler.GenerateJWTWithLabel("forget-password", expired))
}
