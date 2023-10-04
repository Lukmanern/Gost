package main

import (
	"github.com/Lukmanern/gost/application"
	"github.com/Lukmanern/gost/internal/env"
)

func main() {
	env.ReadConfig("./.env")
	config := env.Configuration()
	config.ShowConfig()

	application.RunApp()
}
