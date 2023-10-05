package hash

import (
	"golang.org/x/crypto/bcrypt"
)

// Generate generates hash of data
func Generate(data string) (string, error) {
	pwd := []byte(data)
	pwd, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	response := string(pwd[:])
	return response, nil
}

// Verify verifies password with hashedPassword
func Verify(hashedPassword string, password string) (bool, error) {
	hshPwd := []byte(hashedPassword)
	pwd := []byte(password)
	err := bcrypt.CompareHashAndPassword(hshPwd, pwd)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
