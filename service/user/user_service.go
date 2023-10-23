// don't use this for production

package service

import (
	"context"
	"errors"
	"sync"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gorm.io/gorm"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/hash"
	repository "github.com/Lukmanern/gost/repository/user"
)

type UserService interface {
	Create(ctx context.Context, user model.UserCreate) (id int, err error)
	GetByID(ctx context.Context, id int) (user *model.UserResponse, err error)
	GetByEmail(ctx context.Context, email string) (user *model.UserResponse, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []model.UserResponse, total int, err error)
	Update(ctx context.Context, user model.UserProfileUpdate) (err error)
	Delete(ctx context.Context, id int) (err error)
}

type UserServiceImpl struct {
	repository repository.UserRepository
}

var (
	userService     *UserServiceImpl
	userServiceOnce sync.Once
)

func NewUserService() UserService {
	userServiceOnce.Do(func() {
		userService = &UserServiceImpl{
			repository: repository.NewUserRepository(),
		}
	})

	return userService
}

func (svc UserServiceImpl) Create(ctx context.Context, user model.UserCreate) (id int, err error) {
	userCheck, getErr := svc.GetByEmail(ctx, user.Email)
	if getErr == nil || userCheck != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "email has been used")
	}

	passwordHashed, hashErr := hash.Generate(user.Password)
	if hashErr != nil {
		return 0, errors.New("something failed while hashing data, please try again")
	}

	userEntity := entity.User{
		Name:     cases.Title(language.Und).String(user.Name),
		Email:    user.Email,
		Password: passwordHashed,
	}
	userEntity.SetTimes()

	roleID := entity.USER
	if user.IsAdmin {
		roleID = entity.ADMIN
	}

	id, err = svc.repository.Create(ctx, userEntity, roleID)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (svc UserServiceImpl) GetByID(ctx context.Context, id int) (user *model.UserResponse, err error) {
	userEntity, err := svc.repository.GetByID(ctx, id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return nil, err
	}
	user = &model.UserResponse{
		ID:    userEntity.ID,
		Name:  userEntity.Name,
		Email: userEntity.Email,
	}

	return user, nil
}

func (svc UserServiceImpl) GetByEmail(ctx context.Context, email string) (user *model.UserResponse, err error) {
	userEntity, err := svc.repository.GetByEmail(ctx, email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return nil, err
	}
	user = &model.UserResponse{
		ID:    userEntity.ID,
		Name:  userEntity.Name,
		Email: userEntity.Email,
	}

	return user, nil
}

func (svc UserServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []model.UserResponse, total int, err error) {
	userEntities, total, err := svc.repository.GetAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	users = []model.UserResponse{}
	for _, userEntity := range userEntities {
		newUserResponse := model.UserResponse{
			ID:    userEntity.ID,
			Name:  userEntity.Name,
			Email: userEntity.Email,
		}

		users = append(users, newUserResponse)
	}

	return users, total, nil
}

func (svc UserServiceImpl) Update(ctx context.Context, user model.UserProfileUpdate) (err error) {
	isUserExist := func() bool {
		getUser, getErr := svc.repository.GetByID(ctx, user.ID)
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

	err = svc.repository.Update(ctx, userEntity)
	if err != nil {
		return err
	}

	return nil
}

func (svc UserServiceImpl) Delete(ctx context.Context, id int) (err error) {
	isUserExist := func() bool {
		getUser, getErr := svc.repository.GetByID(ctx, id)
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

	err = svc.repository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
