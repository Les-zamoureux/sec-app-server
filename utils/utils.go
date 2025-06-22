package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
	"math/rand"
)

var ClientUrl string

func HashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GetCurrentDate() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func PasswordValidator(password string) bool {
	containsUpper := false
	containsLower := false
	containsNumber := false
	containsSpecial := false
	minLength := 8

	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			containsUpper = true
		} else if char >= 'a' && char <= 'z' {
			containsLower = true
		} else if char >= '0' && char <= '9' {
			containsNumber = true
		} else if (char >= '!' && char <= '/') || (char >= ':' && char <= '@') || (char >= '[' && char <= '`') || (char >= '{' && char <= '~') {
			containsSpecial = true
		}
	}
	return containsUpper && containsLower && containsNumber && containsSpecial && len(password) >= minLength
}
