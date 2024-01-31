package service

import (
	"testing"

	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/stretchr/testify/assert"
)

const (
	headerTestName string = "at Email Service Test"
)

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
}

func TestEmailService(t *testing.T) {
	service := NewEmailService()

	// success
	err := service.SendMail("subject", "message", "your.valid.email@email.test")
	assert.Nil(t, err, consts.ShouldNil, headerTestName)

	// failed: invalid email
	err = service.SendMail("subject", "message", "_invalid .email@email.test")
	assert.NotNil(t, err, consts.ShouldNotNil, headerTestName)
}
