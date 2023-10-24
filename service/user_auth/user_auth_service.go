package svc

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	userRepository "github.com/Lukmanern/gost/repository/user"
)

type UserAuthService interface {
	FailedLoginCounter(userIP string, increment bool) (counter int, err error)
	Login(ctx context.Context, user model.UserLogin) (token string, err error)
	Logout(c *fiber.Ctx) (err error)
	ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)
	UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)
	UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)
	MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error)
	Register(ctx context.Context, id int) (profile model.UserProfile, err error)
	DeleteUser(ctx context.Context, id int) (err error)
}

type UserAuthServiceImpl struct {
	userRepository userRepository.UserRepository
	jwtHandler     *middleware.JWTHandler
	redis          *redis.Client
}

var (
	userAuthService     *UserAuthServiceImpl
	userAuthServiceOnce sync.Once
)

func NewUserAuthService() UserAuthService {
	userAuthServiceOnce.Do(func() {
		userAuthService = &UserAuthServiceImpl{
			userRepository: userRepository.NewUserRepository(),
			jwtHandler:     middleware.NewJWTHandler(),
			redis:          connector.LoadRedisDatabase(),
		}
	})

	return userAuthService
}

func (svc UserAuthServiceImpl) FailedLoginCounter(userIP string, increment bool) (counter int, err error) {
	key := "failed-login-" + userIP
	getStatus := svc.redis.Get(key)
	counter, _ = strconv.Atoi(getStatus.Val())
	if increment {
		counter++
		setStatus := svc.redis.Set(key, counter, 50*time.Minute)
		if setStatus.Err() != nil {
			return 0, errors.New("storing data to redis")
		}
	}

	return counter, nil
}

func (svc UserAuthServiceImpl) Login(ctx context.Context, user model.UserLogin) (token string, err error) {
	userEntity, err := svc.userRepository.GetByEmail(ctx, user.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return "", err
	}
	if userEntity == nil {
		return "", fiber.NewError(fiber.StatusNotFound, "data not found")
	}
	if !userEntity.IsActive {
		errMsg := "Your account is already exist in our system. But it's still inactive, please check Your email inbox to activated-it"
		return "", fiber.NewError(fiber.StatusBadRequest, errMsg)
	}

	res, verfiryErr := hash.Verify(userEntity.Password, user.Password)
	if verfiryErr != nil {
		return "", verfiryErr
	}
	if !res {
		return "", fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}

	userRole := userEntity.Roles[0]
	permissionMapID := make(rbac.PermissionMap, 0)
	for _, permission := range userRole.Permissions {
		permissionMapID[uint8(permission.ID)] = 0b_0001
	}
	config := env.Configuration()
	expired := time.Now().Add(config.AppAccessTokenTTL)
	token, generetaErr := svc.jwtHandler.GenerateJWT(userEntity.ID, user.Email, userRole.Name, permissionMapID, expired)
	if generetaErr != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "system error while generating token ("+generetaErr.Error()+")")
	}
	if len(token) > 2800 {
		return "", errors.New("token is too large, more than 2800 characters (too large for http header)")
	}

	return token, nil
}

func (svc UserAuthServiceImpl) Logout(c *fiber.Ctx) (err error) {
	err = svc.jwtHandler.InvalidateToken(c)
	if err != nil {
		return errors.New("problem invalidating token")
	}

	return nil
}

func (svc UserAuthServiceImpl) ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error) {
	return nil
}

func (svc UserAuthServiceImpl) UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error) {
	userCheck, err := svc.userRepository.GetByID(ctx, user.ID)
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

	updateErr := svc.userRepository.UpdatePassword(ctx, userCheck.ID, newPasswordHashed)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (svc UserAuthServiceImpl) MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error) {
	user, err := svc.userRepository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return profile, fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return profile, err
	}
	if user == nil {
		return profile, fiber.NewError(fiber.StatusInternalServerError, "error while checking user")
	}

	userRoles := entity.Role{}
	if len(user.Roles) > 0 {
		userRoles = user.Roles[0]
	}
	profile = model.UserProfile{
		Name:  user.Name,
		Email: user.Email,
		Role:  userRoles,
	}

	return profile, nil
}

func (svc UserAuthServiceImpl) UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error) {
	isUserExist := func() bool {
		getUser, getErr := svc.userRepository.GetByID(ctx, user.ID)
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
		Name: cases.Title(language.Und).String(user.Name),
	}
	userEntity.SetUpdateTime()

	err = svc.userRepository.Update(ctx, userEntity)
	if err != nil {
		return err
	}
	return nil
}

func (svc UserAuthServiceImpl) Register(ctx context.Context, id int) (profile model.UserProfile, err error) {
	// Name: cases.Title(language.Und).String(user.Name),
	// Title Casing string
	user, err := svc.userRepository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return profile, fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		return profile, err
	}
	if user == nil {
		return profile, fiber.NewError(fiber.StatusInternalServerError, "error while checking user")
	}

	userRoles := entity.Role{}
	if len(user.Roles) > 0 {
		userRoles = user.Roles[0]
	}
	profile = model.UserProfile{
		Name:  user.Name,
		Email: user.Email,
		Role:  userRoles,
	}

	return profile, nil
}

func (svc UserAuthServiceImpl) DeleteUser(ctx context.Context, id int) (err error) {
	return
}
