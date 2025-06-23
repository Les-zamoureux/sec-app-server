package controller

import (
	"fmt"
	"sec-app-server/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret_jwt_key")

type Credentials struct {
	MailOrUsername string `json:"email"`
	Password       string `json:"password"`
}

func EncodeJWT(mail string) (string, error) {
	fmt.Println(mail)
	isUserAdmin, err := model.IsUserAdmin(mail)
	fmt.Println(isUserAdmin)
	if err != nil {
		fmt.Println("bug here : ", err)
		return "", err
	}

	var role string

	if isUserAdmin {
		role = "admin"
	} else {
		role = "user"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"mail": mail,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"role": role,
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
		username := claims["mail"].(string)
		isUserAdmin, err := model.IsUserAdmin(username)
		fmt.Println("userad", isUserAdmin, username)
		if err != nil {
			return nil, false, err
		}
		return token, isUserAdmin, nil
	}

	return nil, false, fmt.Errorf("invalid token or claims")
}

func GetUserEmailFromGinContext(c *gin.Context) (string, error) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		return "", fmt.Errorf("authorization header is missing")
	}

	tokenString = tokenString[len("Bearer "):] // Remove "Bearer " prefix

	token, isUserAdmin, err := DecodeJWT(tokenString)
	if err != nil {
		return "", err
	}

	if !isUserAdmin {
		return "", fmt.Errorf("user is not an admin")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["mail"].(string)
		return username, nil
	}

	return "", fmt.Errorf("invalid token or claims")
}
