package application

import (
	"net/http"
	"testing"
	"time"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/gofiber/fiber/v2"
)

func init() {
	env.ReadConfig("./../.env")
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
		// user dev
		{"http://localhost:9009/user/dev/99999999", http.StatusNotFound},
		{"http://localhost:9009/user/dev/0", http.StatusBadRequest},
		{"http://localhost:9009/user/dev/-1", http.StatusBadRequest},
		{"http://localhost:9009/user/dev/stringID", http.StatusBadRequest},
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
