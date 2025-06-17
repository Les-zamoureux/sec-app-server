package controller

import (
	"fmt"
	"sec-app-server/model"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_jwt_key")

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func EncodeJWT(username string) (string, error) {
	isUserAdmin, err := model.IsUserAdmin(username)
	if err != nil {
		return "", err
	}

	var role string

	if isUserAdmin {
		role = "admin"
	} else {
		role = "user"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"role":     role,
	})
	return token.SignedString(jwtKey)
}

func DecodeJWT(tokenString string) (*jwt.Token, bool, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method : %s", "ValidationErrorSignatureInvalid")
		}
		return jwtKey, nil
	})

	if err != nil {
		return nil, false, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		isUserAdmin, err := model.IsUserAdmin(username)
		if err != nil {
			return nil, false, err
		}
		return token, isUserAdmin, nil
	}

	return nil, false,  fmt.Errorf("invalid token or claims")
}
