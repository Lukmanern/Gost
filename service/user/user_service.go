package service

import (
	"context"
	"errors"
	"sync"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/hash"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserService interface {
	Create(ctx context.Context, user model.UserCreate) (id int, err error)
	GetByID(ctx context.Context, id int) (user *model.UserResponse, err error)
	GetByEmail(ctx context.Context, email string) (user *model.UserResponse, err error)
	GetAll(ctx context.Context, filter base.RequestGetAll) (users []model.UserResponse, total int, err error)
	Update(ctx context.Context, user model.UserUpdate) (err error)
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

func (service UserServiceImpl) Create(ctx context.Context, user model.UserCreate) (id int, err error) {
	userCheck, err := service.GetByEmail(ctx, user.Email)
	if err == nil || userCheck != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "email has been used")
	}

	passwordHashed, hashErr := hash.Generate(user.Password)
	if hashErr != nil {
		return 0, errors.New("something failed while hashing data, please try again")
	}

	userEntity := entity.User{
		Name:     user.Name,
		Email:    user.Email,
		Password: passwordHashed,
	}
	userEntity.SetTimes()

	id, err = service.repository.Create(ctx, userEntity)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (service UserServiceImpl) GetByID(ctx context.Context, id int) (user *model.UserResponse, err error) {
	userEntity, err := service.repository.GetByID(ctx, id)
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

func (service UserServiceImpl) GetByEmail(ctx context.Context, email string) (user *model.UserResponse, err error) {
	userEntity, err := service.repository.GetByEmail(ctx, email)
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

func (service UserServiceImpl) GetAll(ctx context.Context, filter base.RequestGetAll) (users []model.UserResponse, total int, err error) {
	userEntities, total, err := service.repository.GetAll(ctx, filter)
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

func (service UserServiceImpl) Update(ctx context.Context, user model.UserUpdate) (err error) {
	isUserExist := func() bool {
		getUser, getErr := service.GetByID(ctx, user.ID)
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

	err = service.repository.Update(ctx, userEntity)
	if err != nil {
		return err
	}

	return nil
}

func (service UserServiceImpl) Delete(ctx context.Context, id int) (err error) {
	isUserExist := func() bool {
		getUser, getErr := service.GetByID(ctx, id)
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

	err = service.repository.Delete(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
