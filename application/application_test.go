package application

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/middleware"
	"github.com/Lukmanern/gost/internal/rbac"
	repository "github.com/Lukmanern/gost/repository/user"
	"github.com/gofiber/fiber/v2"
)

var (
	adminRoleID int = 1
	userRoleID  int = 2

	jwtHandler *middleware.JWTHandler
	idHashMap  rbac.PermissionMap
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
	userRepo = repository.NewUserRepository()
	ctx = context.Background()
}

func createUser(role_id int) *entity.User {
	// create new user
	// with admin role
	code := "code"
	createdAt := timeNow.Add(-5 * time.Minute)
	userEntt = entity.User{
		Name:             "name",
		Email:            helper.RandomEmails(2)[0],
		Password:         "password",
		VerificationCode: &code,
		ActivatedAt:      &timeNow,
		TimeFields: base.TimeFields{
			CreatedAt: &createdAt,
			UpdatedAt: &createdAt,
		},
	}
	id, err := userRepo.Create(ctx, userEntt, role_id)
	if err != nil {
		panic("failed to create new user at application/application_test.go : " + err.Error())
	}
	userEntt.ID = id
	return &userEntt
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

func TestRunApp_NOT_FOUND(t *testing.T) {
	go RunApp()
	time.Sleep(5 * time.Second)
	testCases := []struct {
		HTTPMethod   string
		URL          string
		ExpectedCode int
		ExpectedBody string
	}{
		{
			HTTPMethod:   "GET",
			URL:          "http://localhost:9009/not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot GET /not-found-path"}`,
		},
		{
			HTTPMethod:   "POST",
			URL:          "http://localhost:9009/not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot POST /not-found-path"}`,
		},
		{
			HTTPMethod:   "PUT",
			URL:          "http://localhost:9009/not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot PUT /not-found-path"}`,
		},
		{
			HTTPMethod:   "DELETE",
			URL:          "http://localhost:9009/not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot DELETE /not-found-path"}`,
		},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest(tc.HTTPMethod, tc.URL, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
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
		if tc.ExpectedBody != "" {
			if responseStr != tc.ExpectedBody {
				t.Errorf("URL : "+tc.URL+" :: Expected response body '%s', got '%s'", tc.ExpectedBody, responseStr)
			}
		}
		if tc.ExpectedCode == http.StatusNoContent && responseStr != tc.ExpectedBody {
			t.Error("should equal to void-string")
		}
	}
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

func TestRunApp_USER_ROUTE(t *testing.T) {
	// get user
	user := createUser(adminRoleID)
	defer func() {
		deleteUser(user.ID)
	}()

	// get user with role and permission
	getUserByID, getErr := userRepo.GetByID(ctx, user.ID)
	if getErr != nil {
		t.Error("should not error :", getErr)
	}

	userRole := getUserByID.Roles[0]
	permissionMapID := make(rbac.PermissionMap, 0)
	for _, permission := range userRole.Permissions {
		permissionMapID[uint8(permission.ID)] = 0b_0001
	}
	expAt := timeNow.Add(10 * time.Minute)
	token, generateErr := jwtHandler.GenerateJWT(getUserByID.ID, getUserByID.Email, getUserByID.Roles[0].Name, permissionMapID, expAt)
	if generateErr != nil || token == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	// Start the server.
	go RunApp()
	// Wait for the server to run.
	time.Sleep(5 * time.Second)

	// Define test cases with different endpoints,
	// expected status codes, request bodies,
	// and expected response bodies.
	testCases := []struct {
		HTTPMethod   string
		URL          string
		ExpectedCode int
		ReqBody      []byte
		ExpectedBody string
		AddToken     bool
	}{
		{
			HTTPMethod:   "POST",
			URL:          "http://localhost:9009/user",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			HTTPMethod:   "GET",
			URL:          "http://localhost:9009/user/my-profile",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "", // to long
			AddToken:     true,
		},
		{
			HTTPMethod:   "PUT",
			URL:          "http://localhost:9009/user/profile-update",
			ExpectedCode: http.StatusNoContent,
			ReqBody:      []byte(`{"name": "new-name"}`),
			ExpectedBody: "", // no-content
			AddToken:     true,
		},
		{
			HTTPMethod:   "POST",
			URL:          "http://localhost:9009/user/update-password",
			ExpectedCode: http.StatusBadRequest,
			ReqBody:      []byte(`{"password": "password","new_password":"password00DIF","new_password_confirm": "password00"}`),
			ExpectedBody: "", // no-content
			AddToken:     true,
		},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest(tc.HTTPMethod, tc.URL, bytes.NewBuffer(tc.ReqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
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
		if tc.ExpectedBody != "" {
			if responseStr != tc.ExpectedBody {
				t.Errorf("URL : "+tc.URL+" :: Expected response body '%s', got '%s'", tc.ExpectedBody, responseStr)
			}
		}
		if tc.ExpectedCode == http.StatusNoContent && responseStr != tc.ExpectedBody {
			t.Error("should equal to void-string")
		}
	}
}

func TestRunApp_RBAC_TEST(t *testing.T) {
	// get user
	user := createUser(userRoleID)
	defer func() {
		deleteUser(user.ID)
	}()

	// get user with role and permission
	getUserByID, getErr := userRepo.GetByID(ctx, user.ID)
	if getErr != nil {
		t.Error("should not error :", getErr)
	}

	userRole := getUserByID.Roles[0]
	permissionMapID := make(rbac.PermissionMap, 0)
	for _, permission := range userRole.Permissions {
		permissionMapID[uint8(permission.ID)] = 0b_0001
	}
	expAt := timeNow.Add(10 * time.Minute)
	token, generateErr := jwtHandler.GenerateJWT(getUserByID.ID, getUserByID.Email, getUserByID.Roles[0].Name, permissionMapID, expAt)
	if generateErr != nil || token == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	go RunApp()
	time.Sleep(5 * time.Second)
	testCases := []struct {
		AddToken     bool
		HTTPMethod   string
		URL          string
		ExpectedCode int
		ReqBody      []byte
		ExpectedBody string
	}{
		// user with role 'user' should failed to create/ see role and permission
		{
			AddToken:     true,
			HTTPMethod:   "POST",
			URL:          "http://localhost:9009/role",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			AddToken:     true,
			HTTPMethod:   "GET",
			URL:          "http://localhost:9009/role/1",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			AddToken:     true,
			HTTPMethod:   "GET",
			URL:          "http://localhost:9009/role",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			AddToken:     true,
			HTTPMethod:   "POST",
			URL:          "http://localhost:9009/permission",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			AddToken:     true,
			HTTPMethod:   "GET",
			URL:          "http://localhost:9009/permission",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			AddToken:     true,
			HTTPMethod:   "GET",
			URL:          "http://localhost:9009/permission/1",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
	}

	for _, tc := range testCases {
		req, err := http.NewRequest(tc.HTTPMethod, tc.URL, bytes.NewBuffer(tc.ReqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
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
		if tc.ExpectedBody != "" {
			if responseStr != tc.ExpectedBody {
				t.Errorf("URL : "+tc.URL+" :: Expected response body '%s', got '%s'", tc.ExpectedBody, responseStr)
			}
		}
		if tc.ExpectedCode == http.StatusNoContent && responseStr != tc.ExpectedBody {
			t.Error("should equal to void-string")
		}
	}
}

func TestRunApp_MIDDLEWARE_TEST(t *testing.T) {
	// create 2 user: 1 admin and 1 user
	// -> 1 admin -> generateJWT -> run http (looping)
	// -> 1 user -> generateJWT -> run http (looping)

	endpoints := []string{
		// endpoints for admin
		"http://127.0.0:9009/middleware/create-rhp",
		"http://127.0.0:9009/middleware/view-rhp",
		"http://127.0.0:9009/middleware/update-rhp",
		"http://127.0.0:9009/middleware/delete-rhp",
		// endpoints for user
		"http://127.0.0:9009/middleware/create-exmpl",
		"http://127.0.0:9009/middleware/view-exmpl",
		"http://127.0.0:9009/middleware/update-exmpl",
		"http://127.0.0:9009/middleware/delete-exmpl",
	}

	admin := createUser(1)
	user := createUser(2)
	defer func() {
		deleteUser(admin.ID)
		deleteUser(user.ID)
	}()

	// ADMIN
	adminByID, getErr := userRepo.GetByID(ctx, admin.ID)
	if getErr != nil {
		t.Error("should not error :", getErr)
	}

	adminRole := adminByID.Roles[0]
	adminPermissionMapID := make(rbac.PermissionMap, 0)
	for _, permission := range adminRole.Permissions {
		adminPermissionMapID[uint8(permission.ID)] = 0b_0001
	}
	expAt := timeNow.Add(10 * time.Minute)
	adminToken, generateErr := jwtHandler.GenerateJWT(adminByID.ID, adminByID.Email, adminByID.Roles[0].Name, adminPermissionMapID, expAt)
	if generateErr != nil || adminToken == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	// ADMIN
	userByID, getErr := userRepo.GetByID(ctx, user.ID)
	if getErr != nil {
		t.Error("should not error :", getErr)
	}

	userRole := userByID.Roles[0]
	userPermissionMapID := make(rbac.PermissionMap, 0)
	for _, permission := range userRole.Permissions {
		userPermissionMapID[uint8(permission.ID)] = 0b_0001
	}
	userToken, generateErr := jwtHandler.GenerateJWT(userByID.ID, userByID.Email, userByID.Roles[0].Name, userPermissionMapID, expAt)
	if generateErr != nil || userToken == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	for _, endpoint := range endpoints {
		// Test with admin JWT token
		adminClient := &http.Client{}
		reqAdmin, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Errorf("Failed to create request: %v", err)
			continue
		}
		reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
		respAdmin, err := adminClient.Do(reqAdmin)
		if err != nil {
			t.Errorf("Request to %s with admin JWT failed: %v", endpoint, err)
			continue
		}
		defer respAdmin.Body.Close()
		if respAdmin.StatusCode != http.StatusOK {
			t.Errorf("Request to %s with admin JWT returned non-200 status: %d", endpoint, respAdmin.StatusCode)
		}

		// Test with user JWT token
		userClient := &http.Client{}
		reqUser, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			t.Errorf("Failed to create request: %v", err)
			continue
		}
		reqUser.Header.Set("Authorization", "Bearer "+userToken)
		respUser, err := userClient.Do(reqUser)
		if err != nil {
			t.Errorf("Request to %s with user JWT failed: %v", endpoint, err)
			continue
		}
		defer respUser.Body.Close()
		if respUser.StatusCode != http.StatusOK {
			t.Errorf("Request to %s with user JWT returned non-200 status: %d", endpoint, respUser.StatusCode)
		}
	}
}
