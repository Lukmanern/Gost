package main

import (
	"log"

	"github.com/Lukmanern/gost/application"
	"github.com/Lukmanern/gost/internal/env"
)

func main() {
	env.ReadConfig("./.env")
	c := env.Configuration()
	dbURI := c.GetDatabaseURI()
	privKey := c.GetPrivateKey()
	pubKey := c.GetPublicKey()

	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	application.RunApp()
}
