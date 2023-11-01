package middleware

import (
	"testing"
	"time"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

type GenTokenParams struct {
	ID      int
	Email   string
	Role    string
	Per     map[uint8]uint8
	Exp     time.Time
	wantErr bool
}

var (
	params GenTokenParams
)

func init() {
	filepath := "./../../.env"
	env.ReadConfig(filepath)

	timeNow := time.Now()
	params = GenTokenParams{
		ID:    1,
		Email: helper.RandomEmails(1)[0],
		Role:  "test-role",
		Per: map[uint8]uint8{
			1: 1,
			2: 1,
			3: 1,
			4: 1,
			5: 1,
			6: 1,
			7: 1,
			8: 1,
		},
		Exp:     timeNow.Add(5 * time.Minute),
		wantErr: false,
	}
}

func TestNewJWTHandler(t *testing.T) {
	jwtHandler := NewJWTHandler()
	if jwtHandler.publicKey == nil {
		t.Errorf("Public key parsing should have failed")
	}

	if jwtHandler.privateKey == nil {
		t.Errorf("Private key parsing should have failed")
	}
}

func TestGenerateClaims(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(1, params.Email, params.Role, params.Per, params.Exp)
	if err != nil || token == "" {
		t.Fatal("should not error")
	}

	testCases := []struct {
		token    string
		isResNil bool
	}{
		{
			token:    "",
			isResNil: true,
		},
		{
			token:    token,
			isResNil: false,
		},
	}

	for _, tc := range testCases {
		claims := jwtHandler.GenerateClaims(tc.token)
		if claims == nil && !tc.isResNil {
			t.Error("should not nil")
		}
		if claims != nil && tc.isResNil {
			t.Error("should nil")
		}
	}
}

func TestJWTHandler_InvalidateToken(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	invalidErr1 := jwtHandler.InvalidateToken(c)
	if invalidErr1 != nil {
		t.Error("Should error: Expected error for no token")
	}

	c.Request().Header.Add("Authorization", "Bearer "+token)
	invalidErr2 := jwtHandler.InvalidateToken(c)
	if invalidErr2 != nil {
		t.Error("Expected no error for a valid token, but got an error.")
	}
}

func TestJWTHandler_IsBlacklisted(t *testing.T) {
	jwtHandler := NewJWTHandler()
	cookie, err := jwtHandler.GenerateJWT(1000,
		helper.RandomEmails(1)[0], "example-role",
		params.Per, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Error("generate cookie/token should not error")
	}

	type args struct {
		cookie string
	}
	tests := []struct {
		name string
		j    JWTHandler
		args args
		want bool
	}{
		{
			name: "check : false blacklisted",
			j:    *jwtHandler,
			args: args{cookie: cookie},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.j.IsBlacklisted(tt.args.cookie); got != tt.want {
				t.Errorf("JWTHandler.IsBlacklisted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJWTHandler_IsAuthenticated(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	func() {
		jwtHandler1 := NewJWTHandler()
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		jwtHandler1.IsAuthenticated(c)
		c.Status(fiber.StatusUnauthorized)
		if c.Context().Response.StatusCode() != fiber.StatusUnauthorized {
			t.Error("Expected error for no token")
		}
	}()

	func() {
		defer func() {
			r := recover()
			if r != nil {
				t.Error("should not panic")
			}
		}()
		jwtHandler3 := NewJWTHandler()
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		c.Request().Header.Add("Authorization", " "+token)
		c.Status(fiber.StatusUnauthorized)
		jwtHandler3.IsAuthenticated(c)
		if c.Context().Response.StatusCode() != fiber.StatusUnauthorized {
			t.Error("Expected error for no token")
		}
	}()
}

func TestJWTHandler_IsTokenValid(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	isValid := jwtHandler.IsTokenValid(token)
	assert.True(t, isValid, "Valid token should be considered valid")

	isValid = jwtHandler.IsTokenValid("expiredToken")
	assert.False(t, isValid, "Expired token should be considered invalid")

	isValid = jwtHandler.IsTokenValid("invalidToken")
	assert.False(t, isValid, "Invalid token should be considered invalid")
}

func TestJWTHandler_HasPermission(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Add("Authorization", "Bearer "+token)
	jwtHandler.HasPermission(c, "permission-1")
	if c.Response().Header.StatusCode() != fiber.StatusUnauthorized {
		t.Error("Should unauthorized")
	}
}

func TestJWTHandler_HasRole(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	c.Request().Header.Add("Authorization", "Bearer "+token)
	jwtHandler.HasRole(c, "test-role")
	if c.Response().Header.StatusCode() != fiber.StatusUnauthorized {
		t.Error("Should unauthorized")
	}
}

func TestJWTHandler_CheckHasPermission(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	err2 := jwtHandler.CheckHasPermission("permission-1")
	if err2 == nil {
		t.Error("Should unauthorized")
	}
}

func TestJWTHandler_CheckHasRole(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Role, params.Per, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	err2 := jwtHandler.CheckHasRole("permission-1")
	if err2 == nil {
		t.Error("Should unauthorized")
	}
}
