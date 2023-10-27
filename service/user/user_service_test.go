package service

import (
	"log"
	"testing"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	repository "github.com/Lukmanern/gost/repository/user"
	rbacService "github.com/Lukmanern/gost/service/rbac"
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

	// failed forget password : account is created,
	// but account is inactive
	forgetPassErr := svc.ForgetPassword(ctx, model.UserForgetPassword{Email: modelUserRegis.Email})
	if forgetPassErr == nil {
		t.Error("should error login and token should nil-string")
	}
	fiberErr, ok = forgetPassErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusBadRequest {
			t.Error("should error 400BadReq")
		}
	}

	// failed forget password : account is created,
	// but account is inactive
	resetPasswdErr := svc.ResetPassword(ctx, model.UserResetPassword{Code: "wrongCode"})
	if resetPasswdErr == nil {
		t.Error("should error login and token should nil-string")
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

	// success logout
	cForLogout := helper.NewFiberCtx()
	logoutErr := svc.Logout(cForLogout)
	if logoutErr != nil {
		t.Error("should no error")
	}
}

func Test_FailedRegister(t *testing.T) {
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
		RoleID:   -10, // failed
	}
	userID, regisErr := svc.Register(ctx, modelUserRegis)
	if regisErr == nil || userID != 0 {
		t.Error("should error and id should zero")
	}

	defer func() {
		userRepo.Delete(ctx, userID)
	}()

	verifErr := svc.Verification(ctx, "wrongCode")
	if verifErr == nil {
		t.Error("should error")
	}
	fiberErr, ok := verifErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusNotFound {
			t.Error("should error 404")
		}
	}

	deleteUserErr := svc.DeleteUserByVerification(ctx, "wrongCode")
	if deleteUserErr == nil {
		t.Error("should error")
	}
	fiberErr, ok = deleteUserErr.(*fiber.Error)
	if ok {
		if fiberErr.Code != fiber.StatusNotFound {
			t.Error("should error 404")
		}
	}

	// failed login
	_, loginErr := svc.Login(ctx, model.UserLogin{
		IP: "123.1.1.12",
	})
	if loginErr == nil {
		t.Error("should error")
	}

	forgetErr := svc.ForgetPassword(ctx, model.UserForgetPassword{Email: "wrong_email@gost.project"})
	if forgetErr == nil {
		t.Error("should error")
	}

	verifyErr := svc.ResetPassword(ctx, model.UserResetPassword{Code: "wrong-code"})
	if verifyErr == nil {
		t.Error("should error")
	}

	updatePasswdErr := svc.UpdatePassword(ctx, model.UserPasswordUpdate{ID: -1})
	if updatePasswdErr == nil {
		t.Error("should error")
	}

	_, getErr := svc.MyProfile(ctx, -10)
	if getErr == nil {
		t.Error("should error")
	}
}

func Test_Banned_IP_Address(t *testing.T) {
	permSvc := rbacService.NewPermissionService()
	roleSvc := rbacService.NewRoleService(permSvc)
	svc := NewUserService(roleSvc)
	c := helper.NewFiberCtx()
	ctx := c.Context()
	if svc == nil || ctx == nil {
		t.Error("should not nil")
	}

	for i := 1; i <= 15; i++ {
		counter, err := svc.FailedLoginCounter("123.1.1.12", true)
		if err != nil {
			t.Error("should not error")
		}
		if i >= 4 {
			if counter == i {
				t.Error("counter should error")
			}
		}
	}
}

// Register(ctx context.Context, user model.UserRegister) (id int, err error) Done
// Verification(ctx context.Context, verifyCode string) (err error) Done
// DeleteUserByVerification(ctx context.Context, verifyCode string) (err error) Done : half
// FailedLoginCounter(userIP string, increment bool) (counter int, err error) Done
// Login(ctx context.Context, user model.UserLogin) (token string, err error) Done
// Logout(c *fiber.Ctx) (err error) Done
// ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error) Done
// ResetPassword(ctx context.Context, user model.UserResetPassword) (err error) Done
// UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error) Done
// UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error) Done
// MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error) Done

// Todo : add login failed 5 times for banned testing of IP 4-5x times
