// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package controller

import (
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisDatabase()
}

func TestNewDevControllerImpl(t *testing.T) {
	ctr := NewDevControllerImpl()
	c := helper.NewFiberCtx()
	if ctr == nil || c == nil {
		t.Error("should not error")
	}

	pingDbErr := ctr.PingDatabase(c)
	if pingDbErr != nil {
		t.Error("err: ", pingDbErr)
	}

	pingRedisErr := ctr.PingRedis(c)
	if pingRedisErr != nil {
		t.Error("err: ", pingRedisErr)
	}

	panicErr := ctr.Panic(c)
	if panicErr != nil {
		t.Error("err: ", panicErr)
	}

	storingErr := ctr.StoringToRedis(c)
	if storingErr != nil {
		t.Error("err: ", storingErr)
	}

	getErr := ctr.GetFromRedis(c)
	if getErr != nil {
		t.Error("err: ", getErr)
	}
}
