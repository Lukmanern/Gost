package main

import (
	"fmt"

	"github.com/Lukmanern/gost/internal/env"
)

func main() {
	fmt.Println(env.ReadConfig("./.env"))
	// application.RunApp()
}
