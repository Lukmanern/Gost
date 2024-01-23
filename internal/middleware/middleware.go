package middleware

import (
	"crypto/rsa"
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

// JWTHandler struct handles some key and redis connection
// The purpose of this is to handler and checking HTTP Header
// and/or checking is JWT blacklisted or not. See IsBlacklisted func.
type JWTHandler struct {
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
	cache      *redis.Client
}

// Claims struct will be generated as token,contains
// user data like ID, email, role and permissions.
// You can add new field if you want.
type Claims struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTHandler func creates new JwtHandler struct
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

// GenerateJWT func generate new token with expire time for user
func (j *JWTHandler) GenerateJWT(id int, email, role string, expired time.Time) (t string, err error) {
	// Create Claims
	claims := Claims{
		ID:    id,
		Email: email,
		Role:  role,
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

// InvalidateToken func stores (blacklistings) token to redis.
// After storing token in redis, the token is already blacklisted.
// This func is used in Logout feature.
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

// IsBlacklisted func check the token/cookie is blacklisted or not.
func (j JWTHandler) IsBlacklisted(cookie string) bool {
	status := j.cache.Get(cookie)
	val, _ := status.Result()
	return val != ""
}

// IsAuthenticated func extracts token from context (fiber Ctx),
// check is blacklisted or not. And checks the expire time too.
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

// extractToken func extracts token from fiber Ctx.
func extractToken(c *fiber.Ctx) string {
	bearerToken := c.Get(fiber.HeaderAuthorization)
	// Normally Authorization HTTP header.
	onlyToken := strings.Split(bearerToken, " ")
	if len(onlyToken) == 2 {
		return onlyToken[1]
	}

	return ""
}

// GenerateClaims func generates claims struct from jwt.
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

// BuildBitGroups func builds bit-group that can contains
// so much permissions data inside with fast and effective
// with bit manipulations. See the example :
// permissions = {9:10,10:256}
// => read as : bit-group-9th, contains 2 permissions
// => read as : bit-group-10th, contains 8 permissions
// per group contain max 8 permissions sequentially,
// for more You can read in paper (for link, see in readme-md)
func BuildBitGroups(permIDs ...int) map[int]int {
	groups := make(map[int]int)
	for _, id := range permIDs {
		group := (id - 1) / 8
		bitPosition := uint(id - 1 - (group * 8))
		groups[group+1] |= 1 << bitPosition
	}
	return groups
}

// CheckHasPermission func checks if bitGroups (map[int]int)
// contains require permission ID or not
func CheckHasPermission(requirePermID int, userPermissions map[int]int) bool {
	endpointBits := BuildBitGroups(requirePermID)
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

// HasRole func check claims-role equal or not with require role
func (j JWTHandler) HasRole(c *fiber.Ctx, role string) error {
	claims, ok := c.Locals("claims").(*Claims)
	if !ok || role != claims.Role {
		return response.Unauthorized(c)
	}
	return c.Next()
}

// CheckHasRole func is handler/middleware that
// called before the controller for checks the fiber ctx
func (j JWTHandler) CheckHasRole(role string) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return j.HasRole(c, role)
	}
}
