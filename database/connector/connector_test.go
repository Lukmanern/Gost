package connector

import (
	"testing"

	"github.com/Lukmanern/gost/internal/env"
)

func init() {
	filePath := "./../../.env"
	env.ReadConfig(filePath)
}

func TestLoadDatabase(t *testing.T) {
	db := LoadDatabase()
	if db == nil {
		t.Error("db shouldn't nil")
	}
}

func TestLoadRedisDatabase(t *testing.T) {
	rds := LoadRedisDatabase()
	if rds == nil {
		t.Error("rds shouldn't nil")
	}
}
