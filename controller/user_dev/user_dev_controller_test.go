package controller_test

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/Lukmanern/gost/application"
	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/domain/model"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/Lukmanern/gost/internal/rbac"

	controller "github.com/Lukmanern/gost/controller/user_dev"
	service "github.com/Lukmanern/gost/service/user_dev"
)

var (
	userDevService    service.UserDevService
	userDevController controller.UserController
)

func init() {
	// controller\user_dev\user_dev_controller_test.go
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
	r := connector.LoadRedisDatabase()
	r.FlushAll() // clear all key:value in redis

	// dump all permissions into hashMap
	rbac.PermissionNameHashMap = rbac.PermissionNamesHashMap()
	rbac.PermissionHashMap = rbac.PermissionsHashMap()

	userDevService = service.NewUserDevService()
	userDevController = controller.NewUserController(userDevService)

	go application.RunApp()
	time.Sleep(5 * time.Second)
}

func Test_Create(t *testing.T) {
	ctr := userDevController
	if ctr == nil {
		t.Error("should not nil")
	}
	c := helper.NewFiberCtx()
	if ctr == nil || c == nil {
		t.Error("should not error")
	}

	ctx := c.Context()
	if ctx == nil {
		t.Error("should not nil")
	}
	modelUserCreate := model.UserCreate{
		Name:     helper.RandomString(10),
		Email:    helper.RandomEmails(1)[0],
		Password: helper.RandomString(11),
		IsAdmin:  true,
	}
	userDevService.Create(ctx, modelUserCreate)

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

// Create(c *fiber.Ctx) error
// Get(c *fiber.Ctx) error
// GetAll(c *fiber.Ctx) error
// Update(c *fiber.Ctx) error
// Delete(c *fiber.Ctx) error
