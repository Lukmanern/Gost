package service

import (
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	repository "github.com/Lukmanern/gost/repository/user"
	rbacService "github.com/Lukmanern/gost/service/rbac"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")
	c := env.Configuration()
	dbURI := c.GetDatabaseURI()
	privKey := c.GetPrivateKey()
	pubKey := c.GetPublicKey()
	if dbURI == "" || privKey == nil || pubKey == nil {
		log.Fatal("Database URI or keys aren't valid")
	}

	connector.LoadDatabase()
	connector.LoadRedisDatabase()

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()
}

func TestNewUserService(t *testing.T) {
	permSvc := rbacService.NewPermissionService()
	roleSvc := rbacService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	if svc == nil {
		t.Error("should not nil")
	}
}

func Test_SuccessRegister(t *testing.T) {
	permSvc := rbacService.NewPermissionService()
	roleSvc := rbacService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if svc == nil || ctx == nil {
		t.Error("should not nil")
	}

	userRepo := repository.NewUserRepository()
	if userRepo == nil {
		t.Error("should not nil")
	}

	modelUserRegis := model.UserRegister{
		Name:     helper.RandomString(12),
		Email:    helper.RandomEmails(1)[0],
		Password: helper.RandomString(12),
		RoleID:   1, // admin
	}
	userID, regisErr := svc.Register(ctx, modelUserRegis)
	if regisErr != nil || userID < 1 {
		t.Error("should not error and id should more than zero")
	}

	defer func() {
		userRepo.Delete(ctx, userID)
	}()

	userByID, getErr := userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Error("should not error and id should not nil")
	}
	if userByID.Name != cases.Title(language.Und).String(modelUserRegis.Name) ||
		userByID.Email != modelUserRegis.Email ||
		userByID.Roles[0].ID != modelUserRegis.RoleID {
		t.Error("should equal")
	}
	if userByID.VerificationCode == nil {
		t.Error("should not nil")
	}
	if userByID.ActivatedAt != nil {
		t.Error("should nil")
	}

	// failed login : account is created,
	// but account is inactive
	modelUserLogin := model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserRegis.Password,
		IP:       "123.1.1.9",
	}
	token, loginErr := svc.Login(ctx, modelUserLogin)
	if loginErr == nil || token != "" {
		t.Error("should error login and token should nil-string")
	}
	fiberErr, ok := loginErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusBadRequest {
			t.Error("should error 400BadReq")
		}
	}

	vCode := userByID.VerificationCode

	verifErr := svc.Verification(ctx, *vCode)
	if verifErr != nil {
		t.Error("should not nil")
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Error("should not error and id should not nil")
	}
	if userByID.VerificationCode != nil {
		t.Error("should not nil")
	}
	if userByID.ActivatedAt == nil {
		t.Error("should nil")
	}

	// reset value
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserRegis.Password,
		IP:       "123.1.1.9",
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr != nil || token == "" {
		t.Error("should not error login and token should not nil-string")
	}

	jwtHandler := middleware.NewJWTHandler()
	if !jwtHandler.IsTokenValid(token) {
		t.Error("token should valid")
	}
	if jwtHandler.IsBlacklisted(token) {
		t.Error("should not in black-list")
	}

	modelUserForgetPasswd := model.UserForgetPassword{
		Email: modelUserLogin.Email,
	}
	forgetPwErr := svc.ForgetPassword(ctx, modelUserForgetPasswd)
	if forgetPwErr != nil {
		t.Error("should not error")
	}

	// value reset
	userByID = nil
	getErr = nil
	userByID, getErr = userRepo.GetByID(ctx, userID)
	if getErr != nil || userByID == nil {
		t.Error("should not error and id should not nil")
	}
	if userByID.VerificationCode == nil {
		t.Error("should not nil")
	}
	if userByID.ActivatedAt == nil {
		t.Error("should not nil")
	}

	passwd := helper.RandomString(12)
	modelUserResetPasswd := model.UserResetPassword{
		Code:               *userByID.VerificationCode,
		NewPassword:        passwd,
		NewPasswordConfirm: passwd,
	}
	resetErr := svc.ResetPassword(ctx, modelUserResetPasswd)
	if resetErr != nil {
		t.Error("should not error")
	}

	// reset value, login failed
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserRegis.Password,
		IP:       "123.1.1.9",
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr == nil || token != "" {
		t.Error("should error login and token should nil-string")
	}

	// reset value, login success
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserResetPasswd.NewPassword,
		IP:       "123.1.1.9",
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr != nil || token == "" {
		t.Error("should not error login and token should not nil-string")
	}

	passwd = helper.RandomString(14)
	modelUserUpdatePasswd := model.UserPasswordUpdate{
		ID:                 userID,
		OldPassword:        modelUserResetPasswd.NewPassword,
		NewPassword:        passwd,
		NewPasswordConfirm: passwd,
	}
	updatePasswdErr := svc.UpdatePassword(ctx, modelUserUpdatePasswd)
	if updatePasswdErr != nil {
		t.Error("should not error")
	}

	// reset value, login success
	token = ""
	loginErr = nil
	modelUserLogin = model.UserLogin{
		Email:    modelUserRegis.Email,
		Password: modelUserUpdatePasswd.NewPassword,
		IP:       "123.1.1.9",
	}
	token, loginErr = svc.Login(ctx, modelUserLogin)
	if loginErr != nil || token == "" {
		t.Error("should not error login and token should not nil-string")
	}

	modelUserUpdate := model.UserProfileUpdate{
		ID:   userID,
		Name: helper.RandomString(10),
	}
	updateProfileErr := svc.UpdateProfile(ctx, modelUserUpdate)
	if updateProfileErr != nil {
		t.Error("should not error")
	}

	profile, getErr := svc.MyProfile(ctx, userID)
	if getErr != nil {
		t.Error("should not error")
	}
	if profile.Name != cases.Title(language.Und).String(modelUserUpdate.Name) {
		t.Error("should equal")
	}
}

func Test_FailedRegister(t *testing.T) {

}

func Test_IP_Banned(t *testing.T) {

}

// Register(ctx context.Context, user model.UserRegister) (id int, err error)
// Verification(ctx context.Context, verifyCode string) (err error)
// DeleteUserByVerification(ctx context.Context, verifyCode string) (err error)
// FailedLoginCounter(userIP string, increment bool) (counter int, err error)
// Login(ctx context.Context, user model.UserLogin) (token string, err error)
// Logout(c *fiber.Ctx) (err error)
// ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)
// ResetPassword(ctx context.Context, user model.UserResetPassword) (err error)
// UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)
// UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)
// MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error)

// Todo : add login failed 5 times for banned testing of IP 4-5x times
