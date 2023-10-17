package env

import (
	"testing"
)

var (
	config    Config
	validPath = "./../../.env"
)

func TestReadConfigAndConfiguration(t *testing.T) {
	ReadConfig(validPath)
	c := Configuration()

	if c == config {
		t.Error("Expected configuration to be initialized, but it is null")
	}

	if c.AppKey == "" {
		t.Error("AppKey is empty; it should have a valid value")
	}

	if c.AppPort < 1 {
		t.Error("AppPort is less than 1; it should be a positive integer")
	}

	if c.DatabasePort == "" {
		t.Error("DatabasePort is empty; it should have a valid value")
	}

	if c.PublicKey == "" {
		t.Error("PublicKey is empty; it should have a valid value")
	}

	if c.PrivateKey == "" {
		t.Error("PrivateKey is empty; it should have a valid value")
	}

	if c.RedisURI == "" {
		t.Error("RedisURI is empty; it should have a valid value")
	}

	if c.SMTPPort < 1 {
		t.Error("SMTPPort is less than 1; it should be a positive integer")
	}
}

func TestConfig_GetDatabaseURI(t *testing.T) {
	ReadConfig(validPath)
	c := Configuration()
	dbURI := c.GetDatabaseURI()

	if dbURI != c.GetDatabaseURI() {
		t.Error("Expected the same Database URI, but got different URIs")
	}
}

func TestConfig_GetAppInProduction(t *testing.T) {
	ReadConfig(validPath)
	c := Configuration()
	isProd := c.GetAppInProduction()

	if c.AppInProduction != isProd {
		t.Error("Expected AppInProduction to match the configuration, but it doesn't")
	}
}

func TestConfig_GetPublicKeyAndGetPrivateKey(t *testing.T) {
	ReadConfig(validPath)
	c := Configuration()

	pubKey := c.GetPublicKey()
	if pubKey == nil {
		t.Error("Public Key should not be nil")
	}

	privKey := c.GetPrivateKey()
	if privKey == nil {
		t.Error("Private Key should not be nil")
	}
}

func TestConfig_ShowConfig(t *testing.T) {
	ReadConfig(validPath)
	c := Configuration()
	c.ShowConfig()
}
