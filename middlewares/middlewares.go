package middlewares

import (
	"fmt"
	"sec-app-server/controller"
	"strings"

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
