package service

import (
	"log"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/helper"
	repository "github.com/Lukmanern/gost/repository/user"
)

const (
	fileTestName string = "at UserRepoTest"
)

var (
	timeNow time.Time
)

// Register
// Verification
// DeleteUserByVerification
// Login
// ForgetPassword
// ResetPassword
// UpdatePassword
// UpdateProfile
// MyProfile

func init() {
	envFilePath := "./../../.env"
	env.ReadConfig(envFilePath)
	timeNow = time.Now()
}

func createUser(roleID int) entity.User {
	repository := repository.NewUserRepository()
	ctx := helper.NewFiberCtx().Context()
	pwHashed, _ := hash.Generate(helper.RandomString(10))
	user := entity.User{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: pwHashed,
	}
	user.SetCreateTime()
	userID, createErr := repository.Create(ctx, user, roleID)
	if createErr != nil {
		log.Fatal("error while create new user", fileTestName)
	}
	user.ID = userID
	return user
}
