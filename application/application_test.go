package application

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	repository "github.com/Lukmanern/gost/repository/user"
)

var (
	jwtHandler *middleware.JWTHandler
	idHashMap  rbac.PermissionMap
	roleName   string
	roleID     int
	timeNow    time.Time
	userRepo   repository.UserRepository
	ctx        context.Context
	userEntt   entity.User
)

func init() {
	env.ReadConfig("./../.env")

	jwtHandler = middleware.NewJWTHandler()
	idHashMap = rbac.PermissionsHashMap()
	timeNow = time.Now()
	roleName = "admin"
	roleID = 1
	userRepo = repository.NewUserRepository()
	ctx = context.Background()
}

func createUser() entity.User {
	// create new user
	// with admin role
	code := "code"
	createdAt := timeNow.Add(-5 * time.Minute)
	userEntt = entity.User{
		Name:             "name",
		Email:            "email",
		Password:         "password",
		VerificationCode: &code,
		ActivatedAt:      &timeNow,
		TimeFields: base.TimeFields{
			CreatedAt: &createdAt,
			UpdatedAt: &createdAt,
		},
	}
	id, err := userRepo.Create(ctx, userEntt, roleID)
	if err != nil {
		panic("failed to create new user admin at application/application_test.go")
	}
	userEntt.ID = id
	return userEntt
}

func deleteUser(id int) {
	err := userRepo.Delete(ctx, id)
	if err != nil {
		panic("error while deleting user: " + err.Error())
	}
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

func Test_getUserAuthRoutes(t *testing.T) {
	env.ReadConfig("./../.env")
	router := fiber.New()
	getUserRoutes(router)
}

func Test_getUserRoutes(t *testing.T) {
	env.ReadConfig("./../.env")
	router := fiber.New()
	getUserDevRoutes(router)
}

func Test_getEmailRouter(t *testing.T) {
	env.ReadConfig("./../.env")
	router := fiber.New()
	getDevRouter(router)
}

func Test_getRBACAuthRoutes(t *testing.T) {
	env.ReadConfig("./../.env")
	router := fiber.New()
	getRbacRoutes(router)
}

func TestRunApp_HTTP_GET(t *testing.T) {
	// start server
	go RunApp()
	// wait server to run
	time.Sleep(5 * time.Second)

	// Logic todo :
	// Perform HTTP requests to test the running server
	// Catch RESTAPI and StatusCode respon/s
	testCases := []struct {
		URL          string
		ExpectedCode int
	}{
		{"http://localhost:9009/not-found-path", http.StatusNotFound},
		// development user / user management
		{"http://localhost:9009/user-management/99999999", http.StatusNotFound},
		{"http://localhost:9009/user-management/0", http.StatusBadRequest},
		{"http://localhost:9009/user-management/-1", http.StatusBadRequest},
		{"http://localhost:9009/user-management/stringID", http.StatusBadRequest},
		// user
		{"http://localhost:9009/user/my-profile", http.StatusUnauthorized},
		// permission (need auth)
		{"http://localhost:9009/permission/99999999", http.StatusUnauthorized},
		{"http://localhost:9009/permission/0", http.StatusUnauthorized},
		{"http://localhost:9009/permission/-1", http.StatusUnauthorized},
		{"http://localhost:9009/permission/stringID", http.StatusUnauthorized},
		// permission (need auth)
		{"http://localhost:9009/role/99999999", http.StatusUnauthorized},
		{"http://localhost:9009/role/0", http.StatusUnauthorized},
		{"http://localhost:9009/role/-1", http.StatusUnauthorized},
		{"http://localhost:9009/role/stringID", http.StatusUnauthorized},
		// dev
		{"http://localhost:9009/development/ping/db", http.StatusOK},
		{"http://localhost:9009/development/ping/redis", http.StatusOK},
		{"http://localhost:9009/development/panic", http.StatusInternalServerError},
		{"http://localhost:9009/development/new-jwt", http.StatusOK},
		{"http://localhost:9009/development/storing-to-redis", http.StatusCreated},
		{"http://localhost:9009/development/get-from-redis", http.StatusOK},
		// ...
		// Add more test cases here as needed.
	}

	for _, tc := range testCases {
		resp, err := http.Get(tc.URL)
		if err != nil {
			t.Errorf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != tc.ExpectedCode {
			t.Errorf(tc.URL+":: Expected status code %d, got %d", tc.ExpectedCode, resp.StatusCode)
		}
	}
}

func TestRunApp_HTTP_POST(t *testing.T) {
	// get user
	userByID, getErr := userRepo.GetByID(ctx, userEntt.ID)
	if getErr != nil || userByID == nil {
		t.Error("getUserByID :: should not error or not nil")
	}
	expAt := timeNow.Add(10 * time.Minute)
	token, generateErr := jwtHandler.GenerateJWT(userByID.ID, userByID.Email, userByID.Roles[0].Name, map[uint8]uint8{}, expAt)
	if generateErr != nil || token == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	// Start the server.
	go RunApp()
	// Wait for the server to run.
	time.Sleep(5 * time.Second)

	// Define test cases with different URLs,
	// expected status codes, request bodies,
	// and expected response bodies.
	testCases := []struct {
		URL          string
		ExpectedCode int
		ReqBody      []byte
		ExpectedBody string
		AddToken     bool
	}{
		{
			URL:          "http://localhost:9009/not-found-path",
			ExpectedCode: http.StatusNotFound,
			ReqBody:      []byte(`{"key": "value"}`),
			ExpectedBody: `{"message":"Cannot POST /not-found-path"}`,
		},
		{
			URL:          "http://localhost:9009/user",
			ExpectedCode: http.StatusUnauthorized,
			ReqBody:      []byte(`{"user": "test"}`),
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			URL:          "http://localhost:9009/user/my-profile",
			ExpectedCode: http.StatusUnauthorized,
			ReqBody:      []byte(`{"user": "test"}`),
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
	}

	for _, tc := range testCases {

		// Create a request with the Authorization header containing the JWT token.
		req, err := http.NewRequest("POST", tc.URL, bytes.NewBuffer(tc.ReqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		if tc.AddToken {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != tc.ExpectedCode {
			t.Errorf("URL : "+tc.URL+" :: Expected status code %d, got %d", tc.ExpectedCode, resp.StatusCode)
		}

		// Read and verify the response body.
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		}
		responseStr := string(responseBytes)

		if responseStr != tc.ExpectedBody {
			t.Errorf("URL : "+tc.URL+" :: Expected response body '%s', got '%s'", tc.ExpectedBody, responseStr)
		}
	}
}
