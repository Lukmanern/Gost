package middleware

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Lukmanern/gost/database/connector"
	"github.com/Lukmanern/gost/internal/env"
	"github.com/Lukmanern/gost/internal/response"
)

type JWTHandler struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	cache      *redis.Client
}

type Claims struct {
	ID          int         `json:"id"`
	Email       string      `json:"email"`
	Role        string      `json:"role"`
	Permissions map[int]int `json:"permissions"`
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
	newJWTHandler.cache = connector.LoadRedisCache()

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
func (j *JWTHandler) GenerateJWT(id int, email, role string, permissions map[int]int, expired time.Time) (t string, err error) {
	if email == "" || role == "" || len(permissions) < 1 {
		return "", errors.New("email/ role/ permission too short or void")
	}
	// Create Claims
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

func (j JWTHandler) InvalidateToken(c *fiber.Ctx) error {
	cookie := extractToken(c)
	claims := Claims{}
	token, err := jwt.ParseWithClaims(cookie, &claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			message := fmt.Sprintf("unexpected method: %s", jwtToken.Header["alg"])
			return nil, fiber.NewError(fiber.StatusUnauthorized, message)
		}

		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return response.Unauthorized(c)
	}
	status := j.cache.Set(cookie, cookie, time.Until(time.Unix(claims.ExpiresAt.Unix(), 0)))
	if status.Err() != nil {
		return response.Error(c, "problem blacklisting token")
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
		return response.Unauthorized(c)
	}
	claims := Claims{}
	token, err := jwt.ParseWithClaims(cookie, &claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			message := fmt.Sprintf("unexpected method: %s", jwtToken.Header["alg"])
			return nil, fiber.NewError(fiber.StatusUnauthorized, message)
		}

		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return response.Unauthorized(c)
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

func extractToken(c *fiber.Ctx) string {
	bearerToken := c.Get(fiber.HeaderAuthorization)
	// Normally Authorization HTTP header.
	onlyToken := strings.Split(bearerToken, " ")
	if len(onlyToken) == 2 {
		return onlyToken[1]
	}

	return ""
}

func (j JWTHandler) GenerateClaims(cookieToken string) *Claims {
	if j.IsBlacklisted(cookieToken) {
		return nil
	}
	claims := Claims{}
	token, err := jwt.ParseWithClaims(cookieToken, &claims, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			message := fmt.Sprintf("unexpected method: %s", jwtToken.Header["alg"])
			return nil, fiber.NewError(fiber.StatusUnauthorized, message)
		}

		return j.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil
	}
	return &claims
}

func BuildBitGroups(permIDs ...int) map[int]int {
	groups := make(map[int]int)
	for _, id := range permIDs {
		group := (id - 1) / 8
		bitPosition := uint(id - 1 - (group * 8))
		groups[group+1] |= 1 << bitPosition
	}
	return groups
}

func CheckHasPermission(endpointPermID int, userPermissions map[int]int) bool {
	endpointBits := BuildBitGroups(endpointPermID)
	// it seems O(n), but it's actually O(1)
	// because length of $endpointBits is 1
	for key, requiredBits := range endpointBits {
		userBits, ok := userPermissions[key]
		if !ok || requiredBits&userBits == 0 {
			return false
		}
	}
	return true
}

// type PermissionMap = map[uint8]uint8
func (j JWTHandler) HasPermission(c *fiber.Ctx, endpointPermID int) error {
	claims, ok := c.Locals("claims").(*Claims)
	if !ok {
		return response.Unauthorized(c)
	}
	userPermissions := claims.Permissions
	endpointBits := BuildBitGroups(endpointPermID)
	// it seems O(n), but it's actually O(1)
	// because length of $endpointBits is 1
	for key, requiredBits := range endpointBits {
		userBits, ok := userPermissions[key]
		if !ok || requiredBits&userBits == 0 {
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
func (j JWTHandler) CheckHasPermission(endpointPermID int) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return j.HasPermission(c, endpointPermID)
	}
}

// for handler or middleware
func (j JWTHandler) CheckHasRole(role string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return j.HasRole(c, role)
	}
}
