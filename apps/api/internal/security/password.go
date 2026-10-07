package security

import (
	"crypto/rand"
	"fmt"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < 12 {
		return fmt.Errorf("password must contain at least 12 characters")
	}
	if len(password) > 72 {
		return fmt.Errorf("password must not exceed 72 bytes")
	}
	return nil
}

func HashPassword(password string, cost int) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("hash password")
	}
	return string(hash), nil
}

func VerifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func NewDummyPasswordHash(cost int) (string, error) {
	var randomPassword [32]byte
	if _, err := rand.Read(randomPassword[:]); err != nil {
		return "", fmt.Errorf("initialize password verifier")
	}
	hash, err := bcrypt.GenerateFromPassword(randomPassword[:], cost)
	if err != nil {
		return "", fmt.Errorf("initialize password verifier")
	}
	return string(hash), nil
}
