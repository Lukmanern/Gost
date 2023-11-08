package env

import (
	"net/url"
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

func TestConfigGetPublicKeyAndGetPrivateKey(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic")
		}
	}()

	ReadConfig(validPath)
	c := Configuration()

	dbUri := c.GetDatabaseURI()
	if len(dbUri) < 1 {
		t.Error("should more long")
	}

	isProd := c.GetAppInProduction()
	if c.AppInProduction != isProd {
		t.Error("Expected AppInProduction to match the configuration, but it doesn't")
	}

	pubKey := c.GetPublicKey()
	if pubKey == nil {
		t.Error("Public Key should not be nil")
	}

	privKey := c.GetPrivateKey()
	if privKey == nil {
		t.Error("Private Key should not be nil")
	}

	c.ShowConfig()
	c.setAppUrl()
	_, parseUrlErr := url.Parse(c.AppUrl)
	if parseUrlErr != nil {
		t.Error("should not error while parsing url")
	}
}
