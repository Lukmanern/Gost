package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTHandler struct {
	privateKey     *rsa.PrivateKey
	privateKeyOnce sync.Once
}

func NewHWTHandler() *JWTHandler {
	jwtHandler := JWTHandler{}
	jwtHandler.privateKeyOnce.Do(func() {
		rng := rand.Reader
		var err error
		jwtHandler.privateKey, err = rsa.GenerateKey(rng, 2048)
		if err != nil {
			log.Panicf("error while generating rsa-key: %v", err)
		}
		if jwtHandler.privateKey == nil {
			log.Panic("failed generating rsa-key")
		}
	})

	return &jwtHandler
}

// This func used for login.
func (j *JWTHandler) GenerateJWT(id int, email, role string, permissions []string, expired time.Time) (t string, err error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"id":          id,
		"email":       email,
		"role":        role,
		"permissions": permissions,
		"exp":         expired,

		// add Your key-value ...
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t, err = token.SignedString(j.privateKey)
	if err != nil {
		log.Panicf("error while generating JWT : %v", err)
		return "", err
	}
	return t, nil
}

// This func used for forget-password or any.
func (j *JWTHandler) GenerateJWTWithLabel(label string, expired time.Time) (t string, err error) {
	// Create the Claims
	claims := jwt.MapClaims{
		"label": label,
		"exp":   expired,

		// add Your key-value ...
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	t, err = token.SignedString(j.privateKey)
	if err != nil {
		log.Panicf("error while generating JWT : %v", err)
		return "", err
	}
	return t, nil
}
