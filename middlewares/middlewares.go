package middlewares

import (
	"fmt"
	"log"
	"sec-app-server/controller"
	"sec-app-server/db"
	mod "sec-app-server/model"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AdminAuthenticated(handler func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		token, isUserAdmin, err := controller.DecodeJWT(strings.Split(c.GetHeader("Authorization"), " ")[1])
		fmt.Println(isUserAdmin,  token, c.GetHeader("Authorization"))
		if err != nil || token == nil || !isUserAdmin {
			c.JSON(403, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		} else {
			fmt.Println("Ça marche !")
			handler(c)
			return
		}
	}
}

func Authenticated(handler func(c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		token, _, err := controller.DecodeJWT(strings.Split(c.GetHeader("Authorization"), " ")[1])
		if err != nil || token == nil {
			fmt.Println("Error decoding JWT:", err)
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		} else {
			fmt.Println("Ça marche !")
			handler(c)
			return
		}
	}
}

func LogRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := "not connected"
		if c.GetHeader("Authorization") != "" {
			userMail, err := controller.GetUserEmailFromGinContext(c)
			if err == nil {
				user, _ := mod.GetUserByEmailOrUsername(userMail, true)
				userID = user.ID
			}
		}
		c.Next()

		// Enregistre la requête dans la BDD
		if (strings.Contains("GET POST PUT DELETE", c.Request.Method)) {
			_, err := db.DB.Exec(
				"INSERT INTO logs (user_id, method, url, timestamp) VALUES ($1, $2, $3, $4)",
				userID,
				c.Request.Method,
				c.Request.RequestURI,
				time.Now(),
			)
			if err != nil {
				log.Println("Erreur insertion log:", err)
			}
		}
	}
}
