package middleware

import (
	"reflect"
	"testing"
	"time"

	"github.com/Lukmanern/gost/internal/env"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
)

func init() {
	filepath := "./../../.env"
	env.ReadConfig(filepath)
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

func TestJWTHandler_GenerateJWT(t *testing.T) {
	type params struct {
		ID      int
		Email   string
		Role    string
		Per     []string
		Exp     time.Time
		wantErr bool
	}
	timeNow := time.Now()
	paramStruct := []params{
		{
			ID:      1,
			Email:   "test_email@gost.project",
			Role:    "test-role",
			Per:     []string{"permission-1", "permission-2", "permission-3"},
			Exp:     timeNow.Add(60 * time.Hour),
			wantErr: false,
		},
		{
			wantErr: true,
		},
	}
	jwtHandler := NewJWTHandler()
	for _, p := range paramStruct {
		token, err := jwtHandler.GenerateJWT(p.ID, p.Email, p.Role, p.Per, p.Exp)
		if (err != nil) != p.wantErr {
			t.Error("error while generating")
		}
		if token == "" && !p.wantErr {
			t.Error("error token nil")
		}
	}
}

func TestJWTHandler_GenerateJWTWithLabel(t *testing.T) {
	type params struct {
		Email   string
		Exp     time.Time
		wantErr bool
	}
	timeNow := time.Now()
	paramStruct := []params{
		{
			Email:   "Example Label",
			Exp:     timeNow.Add(60 * time.Hour),
			wantErr: false,
		},
		{
			wantErr: true,
		},
	}
	jwtHandler := NewJWTHandler()
	for _, p := range paramStruct {
		token, err := jwtHandler.GenerateJWTWithLabel(p.Email, p.Exp)
		if (err != nil) != p.wantErr {
			t.Error("error while generating")
		}
		if token == "" && !p.wantErr {
			t.Error("error : token void")
		}
	}
}

func TestJWTHandler_InvalidateToken(t *testing.T) {
	type params struct {
		ID      int
		Email   string
		Role    string
		Per     []string
		Exp     time.Time
		wantErr bool
	}
	timeNow := time.Now()
	p := params{
		ID:      1,
		Email:   "test_email@gost.project",
		Role:    "test-role",
		Per:     []string{"permission-1", "permission-2", "permission-3"},
		Exp:     timeNow.Add(60 * time.Hour),
		wantErr: false,
	}
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(p.ID, p.Email, p.Role, p.Per, p.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	invalidErr1 := jwtHandler.InvalidateToken(c)
	if invalidErr1 == nil {
		t.Error("Expected error for no token")
	}

	c.Request().Header.Add("Authorization", "Bearer "+token)
	invalidErr2 := jwtHandler.InvalidateToken(c)
	if invalidErr2 != nil {
		t.Error("Expected no error for a valid token, but got an error.")
	}
}

func TestJWTHandler_IsAuthenticated(t *testing.T) {
	type params struct {
		ID      int
		Email   string
		Role    string
		Per     []string
		Exp     time.Time
		wantErr bool
	}
	timeNow := time.Now()
	p := params{
		ID:      1,
		Email:   "test_email@gost.project",
		Role:    "test-role",
		Per:     []string{"permission-1", "permission-2", "permission-3"},
		Exp:     timeNow.Add(60 * time.Hour),
		wantErr: false,
	}
	jwtHandler := NewJWTHandler()
	token, err := jwtHandler.GenerateJWT(p.ID, p.Email, p.Role, p.Per, p.Exp)
	if err != nil {
		t.Error("error while generating token")
	}
	if token == "" {
		t.Error("error : token void")
	}

	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	_ = jwtHandler.IsAuthenticated(c)
	if c.Context().Response.StatusCode() != fiber.StatusUnauthorized {
		t.Error("Expected error for no token")
	}

	c.Request().Header.Add("Authorization", "Bearer "+token)
	err = jwtHandler.IsAuthenticated(c)
	if err == nil {
		t.Error("Expected an error for no token in the header, but got no error.")
	}
}

func TestJWTHandler_IsTokenValid(t *testing.T) {
	type args struct {
		cookie string
	}
	tests := []struct {
		name string
		j    JWTHandler
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.j.IsTokenValid(tt.args.cookie); got != tt.want {
				t.Errorf("JWTHandler.IsTokenValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJWTHandler_ValidateWithClaim(t *testing.T) {
	type args struct {
		token string
	}
	tests := []struct {
		name      string
		j         JWTHandler
		args      args
		wantClaim jwt.MapClaims
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClaim, err := tt.j.ValidateWithClaim(tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("JWTHandler.ValidateWithClaim() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotClaim, tt.wantClaim) {
				t.Errorf("JWTHandler.ValidateWithClaim() = %v, want %v", gotClaim, tt.wantClaim)
			}
		})
	}
}

func TestJWTHandler_ExtractTokenMetadata(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		j       JWTHandler
		args    args
		want    *Claims
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.j.ExtractTokenMetadata(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("JWTHandler.ExtractTokenMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JWTHandler.ExtractTokenMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractToken(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractToken(tt.args.c); got != tt.want {
				t.Errorf("extractToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJWTHandler_verifyToken(t *testing.T) {
	type args struct {
		c *fiber.Ctx
	}
	tests := []struct {
		name    string
		j       JWTHandler
		args    args
		want    *jwt.Token
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.j.verifyToken(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("JWTHandler.verifyToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JWTHandler.verifyToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJWTHandler_HasPermission(t *testing.T) {
	type args struct {
		c           *fiber.Ctx
		permissions []string
	}
	tests := []struct {
		name    string
		j       JWTHandler
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.j.HasPermission(tt.args.c, tt.args.permissions...); (err != nil) != tt.wantErr {
				t.Errorf("JWTHandler.HasPermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWTHandler_HasRole(t *testing.T) {
	type args struct {
		c     *fiber.Ctx
		roles []string
	}
	tests := []struct {
		name    string
		j       JWTHandler
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.j.HasRole(tt.args.c, tt.args.roles...); (err != nil) != tt.wantErr {
				t.Errorf("JWTHandler.HasRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWTHandler_CheckHasPermission(t *testing.T) {

}

func TestJWTHandler_CheckHasRole(t *testing.T) {

}
