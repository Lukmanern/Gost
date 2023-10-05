package service

import (
	"context"
	"errors"
	"sync"

	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/hash"
	userRepository "github.com/Lukmanern/gost/repository/user"
	"github.com/gofiber/fiber/v2"
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
}

var (
	userAuthService     *UserAuthServiceImpl
	userAuthServiceOnce sync.Once
)

func NewUserAuthService() UserAuthService {
	userAuthServiceOnce.Do(func() {
		userAuthService = &UserAuthServiceImpl{
			userRepository: userRepository.NewUserRepository(),
		}
	})

	return userAuthService
}

func (service UserAuthServiceImpl) Login(ctx context.Context, user model.UserLogin) (token string, err error) {
	userCheck, err := service.userRepository.GetByEmail(ctx, user.Email)
	if err != nil {
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
		return "", fiber.NewError(fiber.StatusInternalServerError, "error while verify password, please try again")
	}

	return "TOKEN-EXAMPLE", nil
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
		return fiber.NewError(fiber.StatusInternalServerError, "error while verify password, please try again")
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
	return nil
}
