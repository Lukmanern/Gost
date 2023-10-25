package svc

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
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
	repository "github.com/Lukmanern/gost/repository/user"
	emailService "github.com/Lukmanern/gost/service/email"
	roleService "github.com/Lukmanern/gost/service/rbac"
)

type UserService interface {
	Register(ctx context.Context, user model.UserRegister) (id int, err error)
	Verification(ctx context.Context, verifyCode string) (err error)
	DeleteUserByVerification(ctx context.Context, verifyCode string) (err error)
	FailedLoginCounter(userIP string, increment bool) (counter int, err error)
	Login(ctx context.Context, user model.UserLogin) (token string, err error)
	Logout(c *fiber.Ctx) (err error)
	ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)
	UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)
	UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)
	MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error)
	DeleteUser(ctx context.Context, id int) (err error)
}

type UserServiceImpl struct {
	repository   repository.UserRepository
	roleService  roleService.RoleService
	emailService emailService.EmailService
	jwtHandler   *middleware.JWTHandler
	redis        *redis.Client
}

var (
	userAuthService     *UserServiceImpl
	userAuthServiceOnce sync.Once
)

func NewUserService(roleService roleService.RoleService) UserService {
	userAuthServiceOnce.Do(func() {
		userAuthService = &UserServiceImpl{
			repository:   repository.NewUserRepository(),
			roleService:  roleService,
			emailService: emailService.NewEmailService(),
			jwtHandler:   middleware.NewJWTHandler(),
			redis:        connector.LoadRedisDatabase(),
		}
	})

	return userAuthService
}

func (svc UserServiceImpl) Register(ctx context.Context, user model.UserRegister) (id int, err error) {
	// search email
	userByEmail, getUserErr := svc.repository.GetByEmail(ctx, user.Email)
	if getUserErr == nil || userByEmail != nil {
		return 0, fiber.NewError(fiber.StatusBadRequest, "email has been used")
	}

	// search role
	roleByID, getRoleErr := svc.roleService.GetByID(ctx, user.RoleID)
	if getRoleErr != nil || roleByID == nil {
		return 0, fiber.NewError(fiber.StatusNotFound, "role not found")
	}

	// create verification code
	var (
		passwordHashed, vCode string
		hashErr               error
		counter               int = 0
	)
	for {
		vCode = randomString(7) + randomString(7) + randomString(7)
		userGetByCode, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
			"verification_code =": vCode,
		})
		if getByCodeErr != nil || userGetByCode == nil {
			break
		}
		counter += 1
		if counter >= 150 {
			return 0, errors.New("failed generating verification code")
		}
	}
	// generate password
	counter = 0
	for {
		passwordHashed, hashErr = hash.Generate(user.Password)
		if hashErr == nil {
			break
		}
		counter += 1
		if counter >= 150 {
			return 0, errors.New("failed hashing user password")
		}
	}

	userEntity := entity.User{
		Name:             cases.Title(language.Und).String(user.Name),
		Email:            user.Email,
		Password:         passwordHashed,
		VerificationCode: &vCode,
		ActivatedAt:      nil,
	}
	userEntity.SetTimes()
	id, err = svc.repository.Create(ctx, userEntity, user.RoleID)
	if err != nil {
		return 0, err
	}

	// sending verify email
	toEmail := []string{user.Email}
	subject := "Gost Project Activation Account"
	message := "Hello, My name is BotGostProject001 from Project Gost: Golang Starter By Lukmanern."
	message += " Your account has already been created but is not yet active. To activate your account,"
	message += " you can click on the Activation Link. If you do not registering for an account or any activity"
	message += " on Project Gost, you can request data deletion by clicking the Link Request Delete."
	message += "\n\n\n\r" // should printed as enter or <br />
	message += ` Activation Link : <a href=http://localhost:9009/user/verification/` + vCode + `"> Verify Now </a> or http://localhost:9009/user/verification/` + vCode
	message += "\n\n\n\r" // should printed as enter or <br />
	message += ` Request Delete Link : <a href=http://localhost:9009/user/request-delete/` + vCode + `"> Verify Now </a> or http://localhost:9009/user/request-delete/` + vCode
	message += "\n\n\n\rThank You, Best Regards BotGostProject001."
	message += " Code : " + vCode

	resMap, sendingErr := svc.emailService.Send(toEmail, subject, message)
	if sendingErr != nil {
		resString, _ := json.Marshal(resMap)
		return 0, errors.New(sendingErr.Error() + " ($data: " +
			string(resString) + "). User Verification Code : " + vCode)
	}

	return id, nil
}

func (svc UserServiceImpl) Verification(ctx context.Context, verifyCode string) (err error) {
	userEntity, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
		"verification_code =": verifyCode,
	})
	if getByCodeErr != nil || userEntity == nil {
		return fiber.NewError(fiber.StatusNotFound, "verification code not found")
	}
	userEntity.ActivatedAccount()
	userEntity.SetUpdateTime()
	updateErr := svc.repository.Update(ctx, *userEntity)
	if updateErr != nil {
		return updateErr
	}
	return nil
}

func (svc UserServiceImpl) DeleteUserByVerification(ctx context.Context, verifyCode string) (err error) {
	userEntity, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
		"verification_code =": verifyCode,
	})
	if getByCodeErr != nil || userEntity == nil {
		return fiber.NewError(fiber.StatusNotFound, "verification code not found")
	}
	deleteErr := svc.repository.Delete(ctx, userEntity.ID)
	if deleteErr != nil {
		return deleteErr
	}
	return nil
}

func (svc UserServiceImpl) FailedLoginCounter(userIP string, increment bool) (counter int, err error) {
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

func (svc UserServiceImpl) Login(ctx context.Context, user model.UserLogin) (token string, err error) {
	userEntity, err := svc.repository.GetByEmail(ctx, user.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return "", err
	}
	if userEntity == nil {
		return "", fiber.NewError(fiber.StatusNotFound, "data not found")
	}

	res, verfiryErr := hash.Verify(userEntity.Password, user.Password)
	if verfiryErr != nil {
		return "", verfiryErr
	}
	if !res {
		return "", fiber.NewError(fiber.StatusBadRequest, "wrong password")
	}
	if userEntity.ActivatedAt == nil {
		errMsg := "Your account is already exist in our system, but it's still inactive, please check Your email inbox to activated-it"
		return "", fiber.NewError(fiber.StatusBadRequest, errMsg)
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

func (svc UserServiceImpl) Logout(c *fiber.Ctx) (err error) {
	err = svc.jwtHandler.InvalidateToken(c)
	if err != nil {
		return errors.New("problem invalidating token")
	}

	return nil
}

func (svc UserServiceImpl) ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error) {
	return nil
}

func (svc UserServiceImpl) UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error) {
	userCheck, err := svc.repository.GetByID(ctx, user.ID)
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

	var (
		hashErr      error
		passwdHashed string
		counter      int
	)
	for {
		passwdHashed, hashErr = hash.Generate(user.NewPassword)
		if hashErr == nil {
			break
		}
		counter += 1
		if counter >= 150 {
			return errors.New("failed hashing user password")
		}
	}

	updateErr := svc.repository.UpdatePassword(ctx, userCheck.ID, passwdHashed)
	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (svc UserServiceImpl) MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error) {
	user, err := svc.repository.GetByID(ctx, id)
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

func (svc UserServiceImpl) UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error) {
	userByID, getErr := svc.repository.GetByID(ctx, user.ID)
	if getErr != nil {
		if getErr == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return err
	}
	if userByID == nil {
		return fiber.NewError(fiber.StatusNotFound, "data not found")
	}

	userEntity := entity.User{
		ID:               user.ID,
		Name:             cases.Title(language.Und).String(user.Name),
		VerificationCode: userByID.VerificationCode,
		ActivatedAt:      userByID.ActivatedAt,
	}
	userEntity.SetUpdateTime()

	err = svc.repository.Update(ctx, userEntity)
	if err != nil {
		return err
	}
	return nil
}

func (svc UserServiceImpl) DeleteUser(ctx context.Context, id int) (err error) {
	return
}

func randomString(n int) string {
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
