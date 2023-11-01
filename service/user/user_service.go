package service

import (
	"context"
	"encoding/json"
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
	"github.com/Lukmanern/gost/internal/helper"
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
	ResetPassword(ctx context.Context, user model.UserResetPassword) (err error)
	UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)
	UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)
	MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error)
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
		passwordHashed, verifCode string
		hashErr                   error
		counter                   int = 0
	)
	for {
		verifCode = ""
		verifCode = helper.RandomString(7) + helper.RandomString(7) + helper.RandomString(7) // total = 21
		userGetByCode, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
			"verification_code =": verifCode,
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
		VerificationCode: &verifCode,
		ActivatedAt:      nil,
	}
	userEntity.SetCreateTimes()
	id, err = svc.repository.Create(ctx, userEntity, user.RoleID)
	if err != nil {
		return 0, err
	}

	toEmail := []string{user.Email}
	subject := "Gost Project Activation Account"
	message := `Hello. This is Your Verification Code : ` + verifCode
	message += "<br /><br />Thank You for registration."

	// resMap, sendingErr := svc.emailService.Send(toEmail, subject, message)
	// if sendingErr != nil {
	// 	resString, _ := json.Marshal(resMap)
	// 	errMsg := fmt.Sprintf("%s ($data: %s). User Verification Code : %s",
	// 		sendingErr, string(resString), verifCode)
	// 	return 0, errors.New(errMsg)
	// }

	sendingErr := svc.emailService.SendMail(toEmail, subject, message)
	if sendingErr != nil {
		return 0, sendingErr
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
	userEntity.SetActivateAccount()
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
	if userEntity.ActivatedAt != nil {
		return fiber.NewError(fiber.StatusBadRequest, "can not delete your data, your account is active")
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
	userEntity, err := svc.repository.GetByEmail(ctx, user.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return err
	}
	if userEntity == nil {
		return fiber.NewError(fiber.StatusNotFound, "data not found")
	}
	if userEntity.ActivatedAt == nil {
		message := "your account has not been activated since register, please check your inbox/ spam mail."
		return fiber.NewError(fiber.StatusBadRequest, message)
	}

	var (
		verifCode string
		counter   int // max retry
	)
	for {
		verifCode = helper.RandomString(7) + helper.RandomString(7) + helper.RandomString(7) // total = 21
		userGetByCode, getByCodeErr := svc.repository.GetByConditions(ctx, map[string]any{
			"verification_code =": verifCode,
		})
		if getByCodeErr != nil || userGetByCode == nil {
			break
		}
		counter += 1
		if counter >= 150 {
			return errors.New("failed generating verification code")
		}
	}
	userEntity.VerificationCode = &verifCode
	userEntity.SetUpdateTime()

	err = svc.repository.Update(ctx, *userEntity)
	if err != nil {
		return err
	}

	// Todo : refactor
	toEmail := []string{user.Email}
	subject := "Gost Project Reset Password"
	message := "Hello, My name is BotGostProject001 from Project Gost: Golang Starter By Lukmanern."
	message += " Your account has already been created but is not yet active. To activate your account,"
	message += " you can click on the Activation Link. If you do not registering for an account or any activity"
	message += " on Project Gost, you can request data deletion by clicking the Link Request Delete."
	message += "\n\n\n\rThank You, Best Regards BotGostProject001."
	message += " Code : " + verifCode

	// Todo : refactor

	resMap, sendingErr := svc.emailService.Send(toEmail, subject, message)
	if sendingErr != nil {
		resString, _ := json.Marshal(resMap)
		return errors.New(sendingErr.Error() + " ($data: " + string(resString) +
			"). User Verification Code : " + verifCode)
	}

	return nil
}

func (svc UserServiceImpl) ResetPassword(ctx context.Context, user model.UserResetPassword) (err error) {
	userByCode, err := svc.repository.GetByConditions(ctx, map[string]any{
		"verification_code =": user.Code,
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return err
	}
	if userByCode == nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	if userByCode.ActivatedAt == nil {
		errMsg := "Your account is already exist in our system, but it's still inactive, please check Your email inbox to activated-it"
		return fiber.NewError(fiber.StatusBadRequest, errMsg)
	}

	var (
		hashErr      error
		passwdHashed string
		counter      int // max retry
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

	userByCode.VerificationCode = nil
	updateErr := svc.repository.Update(ctx, *userByCode)
	if updateErr != nil {
		return updateErr
	}

	updatePasswdErr := svc.repository.UpdatePassword(ctx, userByCode.ID, passwdHashed)
	if updatePasswdErr != nil {
		return updatePasswdErr
	}
	return nil
}

func (svc UserServiceImpl) UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error) {
	userByID, err := svc.repository.GetByID(ctx, user.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "data not found")
		}
		return err
	}
	if userByID == nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	res, verfiryErr := hash.Verify(userByID.Password, user.OldPassword)
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

	updatePwErr := svc.repository.UpdatePassword(ctx, userByID.ID, passwdHashed)
	if updatePwErr != nil {
		return updatePwErr
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

// message := "Hello, My name is BotGostProject001 from Project Gost: Golang Starter By Lukmanern."
// message += " Your account has already been created but is not yet active. To activate your account,"
// message += " you can click on the Activation Link. If you do not registering for an account or any activity"
// message += " on Project Gost, you can request data deletion by clicking the Link Request Delete."
// message += "\n\n\n\rThank You, Best Regards BotGostProject001."
