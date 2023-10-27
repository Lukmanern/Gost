package service

import (
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
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
	// register
	// get by id -> get code
	// verifikasi / Verification -> check verCode is should null
	// try to login -> save(create) JWT
	// forget password -> check verCode is not null
	// Reset Password -> try login
	// Update Password -> try login
	// update profile -> updated or not
	// MyProfile
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
}

// Done Register(ctx context.Context, user model.UserRegister) (id int, err error)
// Done Verification(ctx context.Context, verifyCode string) (err error)
// DeleteUserByVerification(ctx context.Context, verifyCode string) (err error)
// FailedLoginCounter(userIP string, increment bool) (counter int, err error)
// Done Login(ctx context.Context, user model.UserLogin) (token string, err error)
// Logout(c *fiber.Ctx) (err error)
// ForgetPassword(ctx context.Context, user model.UserForgetPassword) (err error)
// ResetPassword(ctx context.Context, user model.UserResetPassword) (err error)
// UpdatePassword(ctx context.Context, user model.UserPasswordUpdate) (err error)
// UpdateProfile(ctx context.Context, user model.UserProfileUpdate) (err error)
// MyProfile(ctx context.Context, id int) (profile model.UserProfile, err error)
