package service

import (
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")
	c := env.Configuration()
	dbURI := c.GetDatabaseURI()
	privKey := c.GetPrivateKey()
	pubKey := c.GetPublicKey()
	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	connector.LoadDatabase()
	connector.LoadRedisDatabase()

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionIDsHashMap()
}

func TestNewEmailServiceAndFuncsGet(t *testing.T) {
	svc := NewEmailService()
	if svc.getAuth() == nil {
		t.Error("should not nil")
	}
	if svc.getSMTPAddr() == "" {
		t.Error("should not nil")
	}
	if svc.getMime() == "" {
		t.Error("should not nil")
	}
}

func TestValidateEmails(t *testing.T) {
	err1 := validateEmails("f", "a")
	if err1 == nil {
		t.Error("should err not nil")
	}

	err2 := validateEmails("validemail@gmail.com")
	if err2 != nil {
		t.Error("should err not nil")
	}

	err3 := validateEmails("validemail@gmail.com", "invalidemail@.gmail.com")
	if err3 == nil {
		t.Error("should err not nil")
	}

	err4 := validateEmails("validemail@gmail.com", "validemail@gmail.com", "invalidemail@gmail.com.")
	if err4 == nil {
		t.Error("should err not nil")
	}
}

func TestTestingHandler(t *testing.T) {
	c := helper.NewFiberCtx()
	svc := NewEmailService()
	err := svc.TestingHandler(c)
	if err != nil {
		t.Error("should not error")
	}
}
