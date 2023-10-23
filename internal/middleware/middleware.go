package middleware

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/rbac"
	"github.com/Lukmanern/gost/internal/response"
	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JWTHandler struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	cache      *redis.Client
}

type Claims struct {
	ID          int                `json:"id"`
	Email       string             `json:"email"`
	Role        string             `json:"role"`
	Permissions rbac.PermissionMap `json:"permissions"`
	Label       *string            `json:"label"`
	jwt.RegisteredClaims
}

func NewJWTHandler() *JWTHandler {
	env.ReadConfig("./.env")
	config := env.Configuration()
	newJWTHandler := JWTHandler{}

	var publicKeyErr error
	newJWTHandler.publicKey, publicKeyErr = jwt.ParseRSAPublicKeyFromPEM([]byte(config.GetPublicKey()))
	if publicKeyErr != nil {
		log.Fatalln("jwt public key parser failed: please check in log file at ./log/log-files")
	}
	var privateKeyErr error
	newJWTHandler.privateKey, privateKeyErr = jwt.ParseRSAPrivateKeyFromPEM([]byte(config.GetPrivateKey()))
	if privateKeyErr != nil {
		log.Fatalln("jwt private key parser failed: please check in log file at ./log/log-files")
	}
	newJWTHandler.cache = connector.LoadRedisDatabase()

	if newJWTHandler.privateKey == nil {
		log.Fatalln("jwt private keys are missed (nil)")
	}
	if newJWTHandler.publicKey == nil {
		log.Fatalln("jwt public keys are missed (nil)")
	}
	if newJWTHandler.cache == nil {
		log.Fatalln("jwt redis cache are missed (nil)")
	}
	return &newJWTHandler
}

// This func used for login.
func (j *JWTHandler) GenerateJWT(id int, email, role string, permissions rbac.PermissionMap, expired time.Time) (t string, err error) {
	if email == "" || role == "" || len(permissions) < 1 {
		return "", errors.New("email/ role/ permission too short or void")
	}
	// Create the Claims
	claims := Claims{
		ID:          id,
		Email:       email,
		Role:        role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: expired},
			NotBefore: &jwt.NumericDate{Time: time.Now()},
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t, err = token.SignedString(j.privateKey)
	if err != nil {
		return "", err
	}
	return t, nil
}

// This func used for forget-password or any.
func (j *JWTHandler) GenerateJWTWithLabel(label string, expired time.Time) (t string, err error) {
	lenLabel := len(label)
	if lenLabel <= 2 || lenLabel > 50 {
		errStr := "label too small or to large (min:3 and max:50)"
		return "", errors.New(errStr)
	}
	// Create the Claims
	claims := Claims{
		Label: &label,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: &jwt.NumericDate{Time: expired},
			NotBefore: &jwt.NumericDate{Time: time.Now()},
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t, err = token.SignedString(j.privateKey)
	if err != nil {
		return "", err
	}
	return t, nil
}

func (j JWTHandler) InvalidateToken(c *fiber.Ctx) error {
	cookie := extractToken(c)
	claims := Claims{}
	token, err := jwt.ParseWithClaims(cookie, &claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("unexpected method: %s", jwtToken.Header["alg"]))
		}

		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}
	status := j.cache.Set(cookie, cookie, time.Until(time.Unix(claims.ExpiresAt.Unix(), 0)))
	if status.Err() != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "problem blacklisting token")
	}
	return nil
}

func (j JWTHandler) IsBlacklisted(cookie string) bool {
	status := j.cache.Get(cookie)

	val, _ := status.Result()

	return val != ""
}

func (j JWTHandler) IsAuthenticated(c *fiber.Ctx) error {
	cookie := extractToken(c)
	if j.IsBlacklisted(cookie) {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}
	claims := Claims{}
	token, err := jwt.ParseWithClaims(cookie, &claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized, fmt.Sprintf("unexpected method: %s", jwtToken.Header["alg"]))
		}

		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthenticated")
	}
	c.Locals("claims", &claims)
	return c.Next()
}

func (j JWTHandler) IsTokenValid(cookie string) bool {
	claims := Claims{}
	token, err := jwt.ParseWithClaims(cookie, &claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fiber.NewError(fiber.StatusUnauthorized)
		}

		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return false
	}
	return true
}

func (j JWTHandler) ValidateWithClaim(token string) (claim jwt.MapClaims, err error) {
	claims := jwt.MapClaims{}
	jwt, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return j.publicKey, nil
	})
	if err != nil || !jwt.Valid {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid token")
	}
	return claims, nil
}

// ExtractTokenMetadata func to extract metadata from JWT.
func (j JWTHandler) ExtractTokenMetadata(c *fiber.Ctx) (*Claims, error) {
	token, err := j.verifyToken(c)
	if err != nil {
		return nil, err
	}

	// Setting and checking token and credentials.
	claims, ok := token.Claims.(*jwt.MapClaims) // Todo : MapClaims -> RegisteredClaims
	if ok && token.Valid {
		condensedClaims := *claims
		// Expires time.
		expires := condensedClaims["exp"]
		expiresTime, ok := expires.(time.Time)
		if !ok {
			return nil, err
		}

		return &Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: &jwt.NumericDate{Time: expiresTime},
				NotBefore: &jwt.NumericDate{Time: time.Now()},
			},
			Email: condensedClaims["email"].(string),
		}, nil
	}

	return nil, err
}

func extractToken(c *fiber.Ctx) string {
	bearerToken := c.Get("Authorization")
	// Normally Authorization HTTP header.
	onlyToken := strings.Split(bearerToken, " ")
	if len(onlyToken) == 2 {
		return onlyToken[1]
	}

	return ""
}

func (j JWTHandler) verifyToken(c *fiber.Ctx) (*jwt.Token, error) {
	tokenString := extractToken(c)
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return j.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	return token, nil
}

// type PermissionMap = map[uint8]uint8
func (j JWTHandler) HasPermission(c *fiber.Ctx, permissions ...string) error {
	claims, ok := c.Locals("claims").(*Claims)
	if !ok {
		return response.Unauthorized(c)
	}
	permissionID := claims.Permissions
	for _, permission := range permissions {
		id, ok1 := rbac.PermissionNameHashMap[permission]
		if !ok1 || id < 1 {
			return response.Unauthorized(c)
		}
		value, ok2 := permissionID[id]
		if !ok2 || value < 1 {
			return response.Unauthorized(c)
		}
	}

	return c.Next()
}

func (j JWTHandler) HasRole(c *fiber.Ctx, role string) error {
	claims, ok := c.Locals("claims").(*Claims)
	if !ok || role != claims.Role {
		return response.Unauthorized(c)
	}
	return c.Next()
}

// for handler or middleware
func (j JWTHandler) CheckHasPermissions(permissions ...string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return j.HasPermission(c, permissions...)
	}
}

// for handler or middleware
func (j JWTHandler) CheckHasRole(role string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return j.HasRole(c, role)
	}
}