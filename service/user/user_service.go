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
	roleRepository "github.com/Lukmanern/gost/repository/role"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserService interface {
	Register(ctx context.Context, data model.UserRegister) (id int, err error)
	Login(ctx context.Context, data model.UserLogin) (token string, err error)
	Logout(c *fiber.Ctx) (err error)

	GetAll(ctx context.Context, filter model.RequestGetAll) (users []model.User, total int, err error)
	MyProfile(ctx context.Context, id int) (profile model.User, err error)

	UpdateProfile(ctx context.Context, data model.UserUpdate) (err error)
	UpdatePassword(ctx context.Context, data model.UserPasswordUpdate) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type UserServiceImpl struct {
	redis      *redis.Client
	jwtHandler *middleware.JWTHandler
	repository repository.UserRepository
	roleRepo   roleRepository.RoleRepository
}

var (
	userSvcImpl     *UserServiceImpl
	userSvcImplOnce sync.Once
)

func NewUserService() UserService {
	userSvcImplOnce.Do(func() {
		userSvcImpl = &UserServiceImpl{
			redis:      connector.LoadRedisCache(),
			jwtHandler: middleware.NewJWTHandler(),
			repository: repository.NewUserRepository(),
			roleRepo:   roleRepository.NewRoleRepository(),
		}
	})
	return userSvcImpl
}

func (svc *UserServiceImpl) Register(ctx context.Context, data model.UserRegister) (id int, err error) {
	_, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr == nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "email already used")
	}

	for _, roleID := range data.RoleIDs {
		enttRole, err := svc.repository.GetByID(ctx, roleID)
		if err == gorm.ErrRecordNotFound {
			return 0, fiber.NewError(fiber.StatusNotFound, consts.NotFound)
		}
		if err != nil || enttRole == nil {
			return 0, errors.New("error while getting role data")
		}
	}

	pwHashed, hashErr := hash.Generate(data.Password)
	if hashErr != nil {
		return 0, errors.New(consts.ErrHashing)
	}

	data.Password = pwHashed
	entityUser := modelRegisterToEntity(data)
	entityUser.SetCreateTime()
	entityUser.ActivatedAt = nil
	id, err = svc.repository.Create(ctx, entityUser, data.RoleIDs)
	if err != nil {
		return 0, errors.New("error while storing user data")
	}
	return id, nil
}

func (svc *UserServiceImpl) Login(ctx context.Context, data model.UserLogin) (token string, err error) {
	user, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr == gorm.ErrRecordNotFound {
		return "", fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return "", errors.New("error while getting user data")
	}

	res, verifyErr := hash.Verify(user.Password, data.Password)
	if verifyErr != nil || !res {
		return "", fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}
	if user.ActivatedAt == nil || user.DeletedAt != nil {
		return "", fiber.NewError(fiber.StatusBadRequest, "your account is inactive, please do activation")
	}

	jwtHandler := middleware.NewJWTHandler()
	expired := time.Now().Add(4 * 24 * time.Hour) // 4 days active
	token, err = jwtHandler.GenerateJWT(user.ID, user.Email, map[string]uint8{}, expired)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return token, nil
}

func (svc *UserServiceImpl) Logout(c *fiber.Ctx) (err error) {
	err = svc.jwtHandler.InvalidateToken(c)
	if err != nil {
		return errors.New("error while logout")
	}
	return nil
}

func (svc *UserServiceImpl) GetAll(ctx context.Context, filter model.RequestGetAll) (users []model.User, total int, err error) {
	entityUsers, total, err := svc.repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	for _, entityUser := range entityUsers {
		users = append(users, entityToResponse(&entityUser))
	}
	return users, total, nil
}

func (svc *UserServiceImpl) MyProfile(ctx context.Context, id int) (profile model.User, err error) {
	user, getErr := svc.repository.GetByID(ctx, id)
	if getErr == gorm.ErrRecordNotFound {
		return model.User{}, fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return model.User{}, errors.New("error while getting user data")
	}
	if user.ActivatedAt == nil || user.DeletedAt != nil {
		return model.User{}, fiber.NewError(fiber.StatusBadRequest, "your account is inactive, please do activation")
	}

	profile = entityToResponse(user)
	return profile, nil
}

func (svc *UserServiceImpl) UpdateProfile(ctx context.Context, data model.UserUpdate) (err error) {
	user, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}
	if user.ActivatedAt == nil || user.DeletedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "your account is inactive, please do activation")
	}

	enttUser := modelUpdateToEntity(data)
	err = svc.repository.Update(ctx, enttUser)
	if err != nil {
		return errors.New("error while updating user data")
	}
	return nil
}

func (svc *UserServiceImpl) UpdatePassword(ctx context.Context, data model.UserPasswordUpdate) (err error) {
	user, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}
	if user.ActivatedAt == nil || user.DeletedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "your account is inactive, please do activation")
	}

	res, verifyErr := hash.Verify(user.Password, data.OldPassword)
	if verifyErr != nil || !res {
		return fiber.NewError(fiber.StatusBadRequest, "wrong password, failed to update")
	}
	pwHashed, hashErr := hash.Generate(data.NewPassword)
	if hashErr != nil {
		return errors.New(consts.ErrHashing)
	}

	updateErr := svc.repository.UpdatePassword(ctx, data.ID, pwHashed)
	if updateErr != nil {
		return errors.New("error while updating user password")
	}
	return nil
}

func (svc *UserServiceImpl) Delete(ctx context.Context, id int) (err error) {
	user, getErr := svc.repository.GetByID(ctx, id)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}

	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return errors.New("error while deleting user password")
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
