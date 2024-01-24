package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/hash"
	"github.com/Lukmanern/gost/internal/middleware"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserService interface {
	GetAll(ctx context.Context, filter model.RequestGetAll) (users []model.User, total int, err error)
	Register(ctx context.Context, data model.UserRegister) (id int, err error)
	Login(ctx context.Context, data model.UserLogin) (token string, err error)
	Logout(c *fiber.Ctx) (err error)
	MyProfile(ctx context.Context, id int) (profile model.User, err error)
	UpdateProfile(ctx context.Context, data model.UserUpdate) (err error)
	UpdatePassword(ctx context.Context, data model.UserPasswordUpdate) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type UserServiceImpl struct {
	jwtHandler *middleware.JWTHandler
	repository repository.UserRepository
	redis      *redis.Client
}

var (
	userSvcImpl     *UserServiceImpl
	userSvcImplOnce sync.Once
)

func NewUserService() UserService {
	userSvcImplOnce.Do(func() {
		userSvcImpl = &UserServiceImpl{
			jwtHandler: middleware.NewJWTHandler(),
			repository: repository.NewUserRepository(),
			redis:      connector.LoadRedisCache(),
		}
	})
	return userSvcImpl
}

func (svc *UserServiceImpl) GetAll(ctx context.Context, filter model.RequestGetAll) (users []model.User, total int, err error) {
	entityUsers, total, getErr := svc.repository.GetAll(ctx, filter)
	if getErr != nil {
		return nil, 0, getErr
	}
	for _, entityUser := range entityUsers {
		users = append(users, entityToResponse(&entityUser))
	}

	return users, total, err
}

func (svc *UserServiceImpl) Register(ctx context.Context, data model.UserRegister) (id int, err error) {
	_, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr == nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "email already used")
	}

	pwHashed, hashErr := hash.Generate(data.Password)
	if hashErr != nil {
		return 0, errors.New(consts.ErrHashing)
	}

	data.Password = pwHashed
	entityUser := modelRegisterToEntity(data)
	entityUser.SetCreateTime()
	id, err = svc.repository.Create(ctx, entityUser, []int{1})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (svc *UserServiceImpl) Login(ctx context.Context, data model.UserLogin) (token string, err error) {
	user, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return "", fiber.NewError(fiber.StatusNotFound, consts.NotFound)
		}
		return "", getErr
	}

	res, verifyErr := hash.Verify(user.Password, data.Password)
	if verifyErr != nil || !res {
		return "", fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}

	jwtHandler := middleware.NewJWTHandler()
	expired := time.Now().Add(4 * 24 * time.Hour)
	token, err = jwtHandler.GenerateJWT(user.ID, user.Email, map[string]uint8{}, expired)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return token, nil
}

func (svc *UserServiceImpl) Logout(c *fiber.Ctx) (err error) {
	err = svc.jwtHandler.InvalidateToken(c)
	if err != nil {
		return err
	}
	return nil
}

func (svc *UserServiceImpl) MyProfile(ctx context.Context, id int) (profile model.User, err error) {
	entityUser, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return profile, fiber.NewError(fiber.StatusNotFound, consts.NotFound)
		}
		return profile, getErr
	}

	profile = entityToResponse(entityUser)
	return profile, nil
}

func (svc *UserServiceImpl) UpdateProfile(ctx context.Context, data model.UserUpdate) (err error) {
	_, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
		}
		return getErr
	}

	user := modelUpdateToEntity(data)
	user.SetUpdateTime()
	err = svc.repository.Update(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

func (svc *UserServiceImpl) UpdatePassword(ctx context.Context, data model.UserPasswordUpdate) (err error) {
	user, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
		}
		return getErr
	}

	res, verifyErr := hash.Verify(user.Password, data.OldPassword)
	if verifyErr != nil || !res {
		return fiber.NewError(fiber.StatusBadRequest, "wrong password, failed to update")
	}

	if user.ActivatedAt != nil || user.DeletedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "your account is inactive")
	}

	pwHashed, hashErr := hash.Generate(data.NewPassword)
	if hashErr != nil {
		return errors.New(consts.ErrHashing)
	}

	updateErr := svc.repository.UpdatePassword(ctx, data.ID, pwHashed)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func (svc *UserServiceImpl) Delete(ctx context.Context, id int) (err error) {
	_, getErr := svc.repository.GetByID(ctx, id)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
		}
		return getErr
	}
	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func modelRegisterToEntity(data model.UserRegister) entity.User {
	return entity.User{
		Name:     data.Name,
		Email:    strings.ToLower(data.Email),
		Password: data.Password,
	}
}

func entityToResponse(data *entity.User) model.User {
	return model.User{
		ID:          data.ID,
		Name:        data.Name,
		Email:       data.Email,
		ActivatedAt: data.ActivatedAt,
	}
}

func modelUpdateToEntity(data model.UserUpdate) entity.User {
	return entity.User{
		ID:   data.ID,
		Name: data.Name,
	}
}
