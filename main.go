package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"sec-app-server/controller"
	"sec-app-server/db"
)

func main() {
	r := gin.Default()

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"prenom_de_merde": "Ian",
		})
	})

	r.POST("/register", func(c *gin.Context) {
		var creds controller.Credentials
		if err := c.ShouldBindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// TODO: Save user to DB (creds.Username, creds.Password)
		c.JSON(http.StatusOK, gin.H{"message": "User registered"})
	})

	r.POST("/connect", func(c *gin.Context) {
		var creds controller.Credentials
		if err := c.ShouldBindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		// TODO: Authenticate user (check creds.Username, creds.Password)
		tokenString, err := controller.EncodeJWT(creds.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	r.POST("/disconnect", func(c *gin.Context) {
		// Invalidate JWT
		c.JSON(http.StatusOK, gin.H{"message": "Disconnected"})
	})

	err := db.InitDB()

	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	db.Test()

	r.Run(":8080")
}
