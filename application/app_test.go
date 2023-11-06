// Don't run test without -p 1
// Please check Makefile file
// or simply just run this : go test -p 1 ./application/...

package application

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	controller "github.com/Lukmanern/gost/controller/development"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	repository "github.com/Lukmanern/gost/repository/user"

	rbacService "github.com/Lukmanern/gost/service/rbac"
	service "github.com/Lukmanern/gost/service/user"
)

type handlerF = func(c *fiber.Ctx) error

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

	t.Run("getUserAuthRoutes", func(t *testing.T) {
		getUserRoutes(router)
	})

	t.Run("getUserRoutes", func(t *testing.T) {
		getUserManagementRoutes(router)
	})
}

func TestDevopmentRouter(t *testing.T) {
	go RunApp()
	time.Sleep(4 * time.Second)

	ctr := controller.NewDevControllerImpl()
	testCases := []struct {
		endpoint string
		handler  handlerF
		payload  any
	}{
		{"ping/db", ctr.PingDatabase, nil},
		{"ping/redis", ctr.PingRedis, nil},
		{"panic", ctr.Panic, nil},
		{"storing-to-redis", ctr.StoringToRedis, nil},
		{"get-from-redis", ctr.GetFromRedis, nil},
		{"auth/test-new-role", ctr.CheckNewRole, nil},
		{"auth/test-new-permission", ctr.CheckNewPermission, nil},
	}

	for _, tc := range testCases {
		jsonObject, err := json.Marshal(&tc.payload)
		if err != nil {
			t.Error("should not error", err.Error())
		}
		req, httpReqErr := http.NewRequest(http.MethodGet, appUrl+"development/"+tc.endpoint, strings.NewReader(string(jsonObject)))
		if httpReqErr != nil {
			t.Fatal("should not nil")
		}
		req.Close = true
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		client := &http.Client{
			Transport: &http.Transport{},
		}
		resp, clientErr := client.Do(req)
		if clientErr != nil {
			t.Fatalf("HTTP request failed: %v", clientErr)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			t.Error("endpoint should exists, but got 404 on: " + tc.endpoint)
		}
	}
}

func TestUserRouter(t *testing.T) {
	ctr := controller.NewDevControllerImpl()
	testCases := []struct {
		endpoint string
		resCode  int
		handler  handlerF
	}{
		{"ping/db", http.StatusOK, ctr.PingDatabase},
		{"ping/redis", http.StatusOK, ctr.PingRedis},
		{"panic", http.StatusInternalServerError, ctr.Panic},
		{"storing-to-redis", http.StatusCreated, ctr.StoringToRedis},
		{"get-from-redis", http.StatusOK, ctr.GetFromRedis},
		{"test-new-role", http.StatusOK, ctr.CheckNewRole},
		{"test-new-permission", http.StatusOK, ctr.CheckNewPermission},
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

func TestUserManagementRouter(t *testing.T) {
	ctr := controller.NewDevControllerImpl()
	testCases := []struct {
		endpoint string
		resCode  int
		handler  handlerF
	}{
		{"ping/db", http.StatusOK, ctr.PingDatabase},
		{"ping/redis", http.StatusOK, ctr.PingRedis},
		{"panic", http.StatusInternalServerError, ctr.Panic},
		{"storing-to-redis", http.StatusCreated, ctr.StoringToRedis},
		{"get-from-redis", http.StatusOK, ctr.GetFromRedis},
		{"test-new-role", http.StatusOK, ctr.CheckNewRole},
		{"test-new-permission", http.StatusOK, ctr.CheckNewPermission},
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
