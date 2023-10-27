package service

import (
	"log"
	"testing"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/rbac"
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
