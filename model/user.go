package model

import (
	"sec-app-server/db"
	"sec-app-server/utils"
)

type User struct {
	ID                int    `json:"id"`
	Username          string `json:"username"`
	Email             string `json:"email"`
	Password          string `json:"password"`
	IsAdmin           bool   `json:"is_admin"`
	VerificationToken string `json:"verification_token"`
	VerificationDate  string `json:"verification_date"`
	CreationDate      string `json:"creation_date"`
}

func AuthenticateUser(username, password string) (*User, error) {
	var user User
	err := db.DB.QueryRow("SELECT * FROM users WHERE username = $1 AND password = $2", utils.HashString(username), utils.HashString(username)).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsAdmin, &user.VerificationToken, &user.VerificationDate, &user.CreationDate)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func IsUserAdmin(username string) (bool, error) {
	var isAdmin bool
	err := db.DB.QueryRow("SELECT is_admin FROM users WHERE username = $1", utils.HashString(username)).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}
