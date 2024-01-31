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
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	roleRepository "github.com/Lukmanern/gost/repository/role"
	repository "github.com/Lukmanern/gost/repository/user"
	service "github.com/Lukmanern/gost/service/email_service"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type UserService interface {
	// no-auth
	Register(ctx context.Context, data model.UserRegister) (id int, err error)
	AccountActivation(ctx context.Context, data model.UserActivation) (err error)
	Login(ctx context.Context, data model.UserLogin) (token string, err error)
	ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)
	ResetPassword(ctx context.Context, user model.UserResetPassword) (err error)
	// auth+admin
	GetAll(ctx context.Context, filter model.RequestGetAll) (users []model.User, total int, err error)
	SoftDelete(ctx context.Context, id int) (err error)
	// auth
	MyProfile(ctx context.Context, id int) (profile model.User, err error)
	Logout(c *fiber.Ctx) (err error)
	UpdateProfile(ctx context.Context, data model.UserUpdate) (err error)
	UpdatePassword(ctx context.Context, data model.UserPasswordUpdate) (err error)
	DeleteAccount(ctx context.Context, data model.UserDeleteAccount) (err error)
}

type UserServiceImpl struct {
	redis        *redis.Client
	jwtHandler   *middleware.JWTHandler
	repository   repository.UserRepository
	roleRepo     roleRepository.RoleRepository
	emailService service.EmailService
}

const (
	KEY_FORGET_PASSWORD    = "-forget-password"
	KEY_ACCOUNT_ACTIVATION = "-account-activation"
)

var (
	userSvcImpl     *UserServiceImpl
	userSvcImplOnce sync.Once
)

func NewUserService() UserService {
	userSvcImplOnce.Do(func() {
		userSvcImpl = &UserServiceImpl{
			redis:        connector.LoadRedisCache(),
			jwtHandler:   middleware.NewJWTHandler(),
			repository:   repository.NewUserRepository(),
			roleRepo:     roleRepository.NewRoleRepository(),
			emailService: service.NewEmailService(),
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
		enttRole, err := svc.roleRepo.GetByID(ctx, roleID)
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

	code := helper.RandomString(32) // verification code
	key := data.Email + KEY_ACCOUNT_ACTIVATION
	exp := time.Hour * 3
	redisStatus := svc.redis.Set(key, code, exp)
	if redisStatus.Err() != nil {
		return id, errors.New("error while storing data to redis")
	}

	subject := "From Gost Project : Successfully User Register"
	message := "This is your verification / activation code."
	message += "This code will expire in 3 hours. <br /><br />Code : " + code
	sendErr := svc.emailService.SendMail(subject, message, strings.ToLower(data.Email))
	if sendErr != nil {
		return id, errors.New("error while sending email confirmation")
	}

	return id, nil
}

func (svc *UserServiceImpl) AccountActivation(ctx context.Context, data model.UserActivation) (err error) {
	user, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}
	if user.ActivatedAt != nil || user.DeletedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "activation failed, account is active or already deleted")
	}

	key := data.Email + KEY_ACCOUNT_ACTIVATION
	redisStatus := svc.redis.Get(key)
	if redisStatus.Err() != nil {
		return errors.New("error while getting data from redis")
	}
	if redisStatus.Val() != data.Code {
		return fiber.NewError(fiber.StatusBadRequest, "verification code isn't match")
	}

	// delete verification code from redis
	svc.redis.Del(key)

	timeNow := time.Now()
	user.ActivatedAt = &timeNow
	err = svc.repository.Update(ctx, *user)
	if err != nil {
		return errors.New("error while updating user data")
	}
	return nil
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
		return "", fiber.NewError(fiber.StatusBadRequest, "account is inactive, please do activation")
	}

	jwtHandler := middleware.NewJWTHandler()
	expired := time.Now().Add(4 * 24 * time.Hour) // 4 days active
	roles := make(map[string]uint8)
	for _, role := range user.Roles {
		roles[role.Name] = 1
	}
	token, err = jwtHandler.GenerateJWT(user.ID, user.Email, roles, expired)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return token, nil
}

func (svc *UserServiceImpl) ForgetPassword(ctx context.Context, data model.UserForgetPassword) (err error) {
	user, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}

	key := data.Email + KEY_FORGET_PASSWORD
	code := helper.RandomString(32)
	exp := time.Hour * 1
	redisStatus := svc.redis.Set(key, code, exp)
	if redisStatus.Err() != nil {
		return errors.New("error while storing data to redis")
	}

	subject := "From Gost Project : Code for Reset Password"
	message := "This code will expire in 1 hours. <br /><br />Code : " + code
	sendErr := svc.emailService.SendMail(subject, message, strings.ToLower(data.Email))
	if sendErr != nil {
		return errors.New("error while sending email confirmation")
	}

	return nil
}

func (svc *UserServiceImpl) ResetPassword(ctx context.Context, data model.UserResetPassword) (err error) {
	user, getErr := svc.repository.GetByEmail(ctx, data.Email)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}
	key := data.Email + KEY_FORGET_PASSWORD
	code := svc.redis.Get(key).Val()
	if code == "" || code != data.Code {
		return fiber.NewError(fiber.StatusNotFound, "verfication code isn't found")
	}

	pwHashed, err := hash.Generate(data.NewPassword)
	if err != nil {
		return errors.New("error while hashing password, please try again")
	}
	err = svc.repository.UpdatePassword(ctx, user.ID, pwHashed)
	if err != nil {
		return errors.New("error while updating password, please try again")
	}

	// delete verification code from redis
	svc.redis.Del(key)
	return nil
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

func (svc *UserServiceImpl) SoftDelete(ctx context.Context, id int) (err error) {
	user, getErr := svc.repository.GetByID(ctx, id)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}

	user.SetDeleteTime()
	err = svc.repository.Update(ctx, *user)
	if err != nil {
		return errors.New("error while updating user data")
	}
	return nil
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
		return model.User{}, fiber.NewError(fiber.StatusBadRequest, "account is inactive, please do activation")
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
		return fiber.NewError(fiber.StatusBadRequest, "account is inactive, please do activation")
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
		return fiber.NewError(fiber.StatusBadRequest, "account is inactive, please do activation")
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

func (svc *UserServiceImpl) DeleteAccount(ctx context.Context, data model.UserDeleteAccount) (err error) {
	user, getErr := svc.repository.GetByID(ctx, data.ID)
	if getErr == gorm.ErrRecordNotFound {
		return fiber.NewError(fiber.StatusNotFound, consts.NotFound)
	}
	if getErr != nil || user == nil {
		return errors.New("error while getting user data")
	}

	res, err := hash.Verify(user.Password, data.Password)
	if err != nil || !res {
		return fiber.NewError(fiber.StatusBadRequest, "wrong password, please try again")
	}

	err = svc.repository.Delete(ctx, data.ID)
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
	roles := make([]string, 0)
	for _, role := range data.Roles {
		roles = append(roles, role.Name)
	}
	return model.User{
		ID:          data.ID,
		Name:        data.Name,
		Email:       data.Email,
		ActivatedAt: data.ActivatedAt,
		DeletedAt:   data.DeletedAt,
		Roles:       roles,
	}
}

func modelUpdateToEntity(data model.UserUpdate) entity.User {
	return entity.User{
		ID:   data.ID,
		Name: data.Name,
	}
}
