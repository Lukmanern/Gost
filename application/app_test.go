// Don't run test per file without -p 1
// or simply run test per func or run
// project test using make test command
// check Makefile file
package application

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/Lukmanern/gost/domain/base"
	"github.com/Lukmanern/gost/domain/entity"
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
	adminRoleID int = 1
	userRoleID  int = 2

	jwtHandler *middleware.JWTHandler
	timeNow    time.Time
	userRepo   repository.UserRepository
	ctx        context.Context
	userEntt   entity.User
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

func createUser(role_id int) *entity.User {
	// create new user
	// with admin role
	code := "code"
	createdAt := timeNow.Add(-5 * time.Minute)
	userEntt = entity.User{
		Name:             "name",
		Email:            helper.RandomEmail(),
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
	getUserManagementRoutes(router)
}

func Test_getEmailRouter(t *testing.T) {
	env.ReadConfig("./../.env")
	router := fiber.New()
	getDevopmentRouter(router)
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
			URL:          appUrl + "not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot GET /not-found-path"}`,
		},
		{
			HTTPMethod:   "POST",
			URL:          appUrl + "not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot POST /not-found-path"}`,
		},
		{
			HTTPMethod:   "PUT",
			URL:          appUrl + "not-found-path",
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"message":"Cannot PUT /not-found-path"}`,
		},
		{
			HTTPMethod:   "DELETE",
			URL:          appUrl + "not-found-path",
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
		{appUrl + "not-found-path", http.StatusNotFound},
		// development user / user management
		{appUrl + "user-management/9999", http.StatusNotFound},
		{appUrl + "user-management/0", http.StatusBadRequest},
		{appUrl + "user-management/-1", http.StatusBadRequest},
		{appUrl + "user-management/stringID", http.StatusBadRequest},
		// user
		{appUrl + "user/my-profile", http.StatusUnauthorized},
		// permission (need auth)
		{appUrl + "permission/9999", http.StatusUnauthorized},
		{appUrl + "permission/0", http.StatusUnauthorized},
		{appUrl + "permission/-1", http.StatusUnauthorized},
		{appUrl + "permission/stringID", http.StatusUnauthorized},
		// permission (need auth)
		{appUrl + "role/9999", http.StatusUnauthorized},
		{appUrl + "role/0", http.StatusUnauthorized},
		{appUrl + "role/-1", http.StatusUnauthorized},
		{appUrl + "role/stringID", http.StatusUnauthorized},
		// dev
		{appUrl + "development/ping/db", http.StatusOK},
		{appUrl + "development/ping/redis", http.StatusOK},
		{appUrl + "development/panic", http.StatusInternalServerError},
		{appUrl + "development/storing-to-redis", http.StatusCreated},
		{appUrl + "development/get-from-redis", http.StatusOK},
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
	permIDs := make([]int, 0)
	for _, perm := range userRole.Permissions {
		permIDs = append(permIDs, perm.ID)
	}
	bitGroups := middleware.BuildBitGroups(permIDs...)
	expAt := timeNow.Add(6 * time.Minute)
	token, generateErr := jwtHandler.GenerateJWT(getUserByID.ID, getUserByID.Email, getUserByID.Roles[0].Name, bitGroups, expAt)
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
			URL:          appUrl + "user",
			ExpectedCode: http.StatusUnauthorized,
			ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		},
		{
			HTTPMethod:   "GET",
			URL:          appUrl + "user/my-profile",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "", // to long
			AddToken:     true,
		},
		{
			HTTPMethod:   "PUT",
			URL:          appUrl + "user/profile-update",
			ExpectedCode: http.StatusNoContent,
			ReqBody:      []byte(`{"name": "new-name"}`),
			ExpectedBody: "", // no-content
			AddToken:     true,
		},
		{
			HTTPMethod:   "POST",
			URL:          appUrl + "user/update-password",
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
	permIDs := make([]int, 0)
	for _, perm := range userRole.Permissions {
		permIDs = append(permIDs, perm.ID)
	}
	bitGroups := middleware.BuildBitGroups(permIDs...)
	expAt := timeNow.Add(6 * time.Minute)
	token, generateErr := jwtHandler.GenerateJWT(getUserByID.ID, getUserByID.Email, getUserByID.Roles[0].Name, bitGroups, expAt)
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
		// {
		// 	AddToken:     true,
		// 	HTTPMethod:   "POST",
		// 	URL:          appUrl + "role",
		// 	ExpectedCode: http.StatusUnauthorized,
		// 	ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		// },
		// {
		// 	AddToken:     true,
		// 	HTTPMethod:   "GET",
		// 	URL:          appUrl + "role/1",
		// 	ExpectedCode: http.StatusUnauthorized,
		// 	ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		// },
		// {
		// 	AddToken:     true,
		// 	HTTPMethod:   "GET",
		// 	URL:          appUrl + "role",
		// 	ExpectedCode: http.StatusUnauthorized,
		// 	ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		// },
		// {
		// 	AddToken:     true,
		// 	HTTPMethod:   "POST",
		// 	URL:          appUrl + "permission",
		// 	ExpectedCode: http.StatusUnauthorized,
		// 	ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		// },
		// {
		// 	AddToken:     true,
		// 	HTTPMethod:   "GET",
		// 	URL:          appUrl + "permission",
		// 	ExpectedCode: http.StatusUnauthorized,
		// 	ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		// },
		// {
		// 	AddToken:     true,
		// 	HTTPMethod:   "GET",
		// 	URL:          appUrl + "permission/1",
		// 	ExpectedCode: http.StatusUnauthorized,
		// 	ExpectedBody: `{"message":"unauthorized","success":false,"data":null}`,
		// },
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

func TestRunApp_MIDDLEWARE_ADMIN_TEST(t *testing.T) {
	adminEndpoints := []string{
		// appUrl + "middleware/create-rhp",
		// appUrl + "middleware/view-rhp",
		// appUrl + "middleware/update-rhp",
		// appUrl + "middleware/delete-rhp",
	}
	userEndpoints := []string{
		// appUrl + "middleware/create-exmpl",
		// appUrl + "middleware/view-exmpl",
		// appUrl + "middleware/update-exmpl",
		// appUrl + "middleware/delete-exmpl",
	}

	admin := createUser(1)
	defer func() {
		deleteUser(admin.ID)
	}()

	// ADMIN
	adminByID, getErr := userRepo.GetByID(ctx, admin.ID)
	if getErr != nil {
		t.Error("should not error :", getErr)
	}

	adminRole := adminByID.Roles[0]
	permIDs := make([]int, 0)
	for _, perm := range adminRole.Permissions {
		permIDs = append(permIDs, perm.ID)
	}
	bitGroups := middleware.BuildBitGroups(permIDs...)
	expAt := timeNow.Add(6 * time.Minute)
	adminToken, generateErr := jwtHandler.GenerateJWT(adminByID.ID, adminByID.Email, adminByID.Roles[0].Name, bitGroups, expAt)
	if generateErr != nil || adminToken == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	go RunApp()
	time.Sleep(5 * time.Second)

	for _, endpoint := range adminEndpoints {
		req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Error("URL : " + endpoint + " :: Expected status 200OK")
		}
		// Read and verify the response body.
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		}
		responseStr := string(responseBytes)
		shouldRespone := `{"message":"success-view-endpoint","success":true,"data":null}`
		if responseStr != shouldRespone {
			t.Error("should be : " + responseStr)
		}
	}
	for _, endpoint := range userEndpoints {
		req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Error("URL : " + endpoint + " :: Expected status 401")
		}
		// Read and verify the response body.
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		}
		responseStr := string(responseBytes)
		shouldRespone := `{"message":"unauthorized","success":false,"data":null}`
		if responseStr != shouldRespone {
			t.Error("should be : " + responseStr)
		}
	}
}

func TestRunApp_MIDDLEWARE_USER_TEST(t *testing.T) {
	adminEndpoints := []string{
		// appUrl + "middleware/create-rhp",
		// appUrl + "middleware/view-rhp",
		// appUrl + "middleware/update-rhp",
		// appUrl + "middleware/delete-rhp",
	}
	userEndpoints := []string{
		// appUrl + "middleware/create-exmpl",
		// appUrl + "middleware/view-exmpl",
		// appUrl + "middleware/update-exmpl",
		// appUrl + "middleware/delete-exmpl",
	}

	user := createUser(2)
	defer func() {
		deleteUser(user.ID)
	}()

	// ADMIN
	userByID, getErr := userRepo.GetByID(ctx, user.ID)
	if getErr != nil {
		t.Error("should not error :", getErr)
	}

	userRole := userByID.Roles[0]
	permIDs := make([]int, 0)
	for _, perm := range userRole.Permissions {
		permIDs = append(permIDs, perm.ID)
	}
	bitGroups := middleware.BuildBitGroups(permIDs...)
	expAt := timeNow.Add(6 * time.Minute)
	userToken, generateErr := jwtHandler.GenerateJWT(userByID.ID, userByID.Email, userByID.Roles[0].Name, bitGroups, expAt)
	if generateErr != nil || userToken == "" {
		t.Error("generateJWT :: should not error or not void string")
	}

	go RunApp()
	time.Sleep(5 * time.Second)

	for _, endpoint := range adminEndpoints {
		req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Error("URL : " + endpoint + " :: Expected status 401")
		}
		// Read and verify the response body.
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		}
		responseStr := string(responseBytes)
		shouldRespone := `{"message":"unauthorized","success":false,"data":null}`
		if responseStr != shouldRespone {
			t.Error("should be : " + responseStr)
		}
	}
	for _, endpoint := range userEndpoints {
		req, err := http.NewRequest("GET", endpoint, bytes.NewBuffer([]byte{}))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Error("URL : " + endpoint + " :: Expected status 200OK")
		}
		// Read and verify the response body.
		responseBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("Failed to read response body: %v", err)
		}
		responseStr := string(responseBytes)
		shouldRespone := `{"message":"success-view-endpoint","success":true,"data":null}`
		if responseStr != shouldRespone {
			t.Error("should be : " + responseStr)
		}
	}
}

func createUserAndToken(roleID int) (int, string) {
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
