package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/jwt"
	userRepository "github.com/Lukmanern/gost/repository/user"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserAuthService interface {
	Login(ctx context.Context, user model.UserLogin) (token string, err error)
	Logout(ctx context.Context) (err error)
	ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)
	UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)
	UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)
}

type UserAuthServiceImpl struct {
	userRepository userRepository.UserRepository
	jwtHandler     *jwt.JWTHandler
}

var (
	userAuthService     *UserAuthServiceImpl
	userAuthServiceOnce sync.Once
)

func NewUserAuthService() UserAuthService {
	userAuthServiceOnce.Do(func() {
		userAuthService = &UserAuthServiceImpl{
			userRepository: userRepository.NewUserRepository(),
			jwtHandler:     jwt.NewJWTHandler(),
		}
	})

	return userAuthService
}

func (service UserAuthServiceImpl) Login(ctx context.Context, user model.UserLogin) (token string, err error) {
	userCheck, err := service.userRepository.GetByEmail(ctx, user.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return "", err
	}
	if userCheck == nil {
		return "", fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	res, verfiryErr := hash.Verify(userCheck.Password, user.Password)
	if verfiryErr != nil {
		return "", verfiryErr
	}
	if !res {
		return "", fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}

	config := env.Configuration()
	expired := time.Now().Add(config.AppAccessTokenTTL)
	token, generetaErr := service.jwtHandler.GenerateJWT(userCheck.ID, user.Email, "role", []string{"permission-1", "permission-2"}, expired)
	if generetaErr != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "system error while generating token, please try again")
	}

	return token, nil
}

func (service UserAuthServiceImpl) Logout(ctx context.Context) (err error) {
	return nil
}

func (service UserAuthServiceImpl) ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error) {
	return nil
}

func (service UserAuthServiceImpl) UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error) {
	userCheck, err := service.userRepository.GetByID(ctx, user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return err
	}
	if userCheck == nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	res, verfiryErr := hash.Verify(userCheck.Password, user.OldPassword)
	if verfiryErr != nil {
		return verfiryErr
	}
	if !res {
		return fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}

	newPasswordHashed, hashErr := hash.Generate(user.NewPassword)
	if hashErr != nil {
		return errors.New("something failed while hashing new data, please try again")
	}

	updateErr := service.userRepository.UpdatePassword(ctx, userCheck.ID, newPasswordHashed)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (service UserAuthServiceImpl) UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error) {
	isUserExist := func() bool {
		getUser, getErr := service.userRepository.GetByID(ctx, user.ID)
		if getErr != nil {
			return false
		}
		if getUser == nil {
			return false
		}

		return true
	}
	if !isUserExist() {
		return fiber.NewError(fiber.StatusNotFound, "data not found")
	}

	userEntity := entity.User{
		ID:   user.ID,
		Name: user.Name,
	}
	userEntity.SetUpdateTime()

	err = service.userRepository.Update(ctx, userEntity)
	if err != nil {
		return err
	}
	return nil
}
