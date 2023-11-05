// Don't run test without -p 1
// Please check Makefile file
// or simply just run this : go test -p 1 ./application/...

package application

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	controller "github.com/Lukmanern/gost/controller/development"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	repository "github.com/Lukmanern/gost/repository/user"
	rbacService "github.com/Lukmanern/gost/service/rbac"
	service "github.com/Lukmanern/gost/service/user"
	"github.com/gofiber/fiber/v2"
)

var (
	jwtHandler *middleware.JWTHandler
	timeNow    time.Time
	userRepo   repository.UserRepository
	ctx        context.Context
	appUrl     string
)

func init() {
	env.ReadConfig("./../.env")
	c := env.Configuration()
	appUrl = c.AppUrl

	jwtHandler = middleware.NewJWTHandler()
	timeNow = time.Now()
	userRepo = repository.NewUserRepository()
	ctx = context.Background()
}

// helper func
func CreateUserAndToken(roleID int) (int, string) {
	permissionService := rbacService.NewPermissionService()
	roleService := rbacService.NewRoleService(permissionService)
	userService := service.NewUserService(roleService)

	userID, regisErr := userService.Register(ctx, model.UserRegister{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmail(),
		Password: helper.RandomString(10),
		RoleID:   roleID,
	})
	if regisErr != nil {
		log.Fatalf("\n\nfailed create user, error: %v\n", regisErr)
	}
	userService.MyProfile(ctx, userID)
	userService.Verification(ctx, "")

	return 0, ""
}

func Test_RunApp(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic", r)
		}
	}()

	go RunApp()
	time.Sleep(3 * time.Second)
}

func Test_app_router(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic : ", r)
		}
	}()

	if router == nil {
		t.Error("Router should not be nil")
	}
	if router.Server() == nil {
		t.Error("Router's server should not be nil")
	}
	if router.Config().ReadBufferSize <= 0 {
		t.Error("Router's ReadBufferSize should be more than 0")
	}
	if router.Config().WriteBufferSize <= 0 {
		t.Error("Router's WriteBufferSize should be more than 0")
	}
	if router.Config().ServerHeader != "" {
		t.Error("Router's ServerHeader should be empty")
	}
	if router.Config().ProxyHeader != "" {
		t.Error("Router's ProxyHeader should be empty")
	}
	setup()
}

func TestDevopmentRouter(t *testing.T) {
	ctr := controller.NewDevControllerImpl()
	testCases := []struct {
		endpoint string
		resCode  int
		handler  func(c *fiber.Ctx) error
	}{
		{"ping/db", http.StatusOK, ctr.PingDatabase},
		{"ping/redis", http.StatusOK, ctr.PingRedis},
		{"panic", http.StatusInternalServerError, ctr.Panic},
		{"storing-to-redis", http.StatusCreated, ctr.StoringToRedis},
		{"get-from-redis", http.StatusOK, ctr.GetFromRedis},
	}

	for _, tc := range testCases {
		app := fiber.New()
		req := httptest.NewRequest(http.MethodGet, "/development/"+tc.endpoint, nil)
		app.Get("/development/"+tc.endpoint, tc.handler)
		res, err := app.Test(req, -1)
		if err != nil {
			t.Fatal("should not error")
		}
		defer res.Body.Close()
		if res.StatusCode != tc.resCode {
			t.Errorf("Expected status code %d, but got %d", tc.resCode, res.StatusCode)
		}
	}
}

func TestRoutes(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Error("should not panic : ", r)
		}
	}()
	router := fiber.New()

	t.Run("getDevelopmentRouter", func(t *testing.T) {
		getDevopmentRouter(router)
	})

	t.Run("getRBACAuthRoutes", func(t *testing.T) {
		getRbacRoutes(router)
	})

	t.Run("getMiddlewareTestingRoute", func(t *testing.T) {
		getMiddlewareTestingRoute(router)
	})

	t.Run("getUserAuthRoutes", func(t *testing.T) {
		getUserRoutes(router)
	})

	t.Run("getUserRoutes", func(t *testing.T) {
		getUserManagementRoutes(router)
	})
}
