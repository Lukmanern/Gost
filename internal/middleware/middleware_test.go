package middleware

import (
	"testing"
	"time"

	"github.com/Lukmanern/gost/internal/consts"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/helper"
	"github.com/gofiber/fiber/v2"
)

type GenTokenParams struct {
	ID      int
	Email   string
	Roles   map[string]uint8
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
		ID:      helper.GenerateRandomID(),
		Email:   helper.RandomEmail(),
		Roles:   map[string]uint8{"test-role": 1},
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
	token, err := jwtHandler.GenerateJWT(1, params.Email, params.Roles, params.Exp)
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

func TestJWTHandlerInvalidateToken(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Roles, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}
	c := helper.NewFiberCtx()
	invalidErr1 := jwtHandler.InvalidateToken(c)
	if invalidErr1 != nil {
		t.Error("Should error: Expected error for no token")
	}

	c.Request().Header.Add(fiber.HeaderAuthorization, "Bearer "+token)
	invalidErr2 := jwtHandler.InvalidateToken(c)
	if invalidErr2 != nil {
		t.Error("Expected no error for a valid token, but got an error.")
	}
}

func TestJWTHandlerIsBlacklisted(t *testing.T) {
	jwtHandler := NewJWTHandler()
	cookie, err := jwtHandler.GenerateJWT(1000,
		helper.RandomEmail(), params.Roles,
		time.Now().Add(1*time.Hour))
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

func TestJWTHandlerIsAuthenticated(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Roles, params.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	func() {
		jwtHandler1 := NewJWTHandler()
		c := helper.NewFiberCtx()
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
				t.Error("should not panic", r)
			}
		}()
		jwtHandler3 := NewJWTHandler()
		c := helper.NewFiberCtx()
		c.Request().Header.Add(fiber.HeaderAuthorization, " "+token)
		c.Status(fiber.StatusUnauthorized)
		jwtHandler3.IsAuthenticated(c)
		if c.Context().Response.StatusCode() != fiber.StatusUnauthorized {
			t.Error("Expected error for no token")
		}
	}()
}

func TestJWTHandlerCheckHasRole(t *testing.T) {
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(params.ID, params.Email, params.Roles, params.Exp)
	if err != nil {
		t.Error("Error while generating token:", err)
	}
	if token == "" {
		t.Error("Error: Token is empty")
	}

	checkErr := jwtHandler.HasRole("role-x-1")
	if checkErr == nil {
		t.Error(consts.Unauthorized)
	}
}
