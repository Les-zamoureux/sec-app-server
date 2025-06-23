package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"sec-app-server/controller"
	"sec-app-server/db"
	mailcontroller "sec-app-server/mail_controller"
	m "sec-app-server/middlewares"
	mod "sec-app-server/model"
	"sec-app-server/utils"
)

func main() {
	r := gin.Default()

	r.Use(m.LogRequest())

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mailcontroller.InitMailSystem()

	utils.ClientUrl = os.Getenv("CLIENT_URL")

	origins := utils.ClientUrl
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{origins},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))

	initUserRoutes(r)
	initProductRoutes(r)
	initFAQRoutes(r)
	initLogsRoutes(r)

	r.Static("/uploads", "./uploads")

	err = db.InitDB()

	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	db.Test()

	r.Run(":8080")
}

func initUserRoutes(r *gin.Engine) {
	r.POST("/register", func(c *gin.Context) {
		var creds struct {
			Username string `json:"username" binding:"required"`
			Mail     string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if !utils.PasswordValidator(creds.Password) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters long, contain at least one uppercase letter, one lowercase letter, one number, and one special character"})
			return
		}

		usernameExists, emailExists, err := mod.CheckUserExists(creds.Username, creds.Mail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
			return
		}
		if usernameExists && emailExists {
			c.JSON(http.StatusConflict, gin.H{"error": "already-used:username&email"})
			return
		}
		if usernameExists {
			c.JSON(http.StatusConflict, gin.H{"error": "already-used:username"})
			return
		}
		if emailExists {
			c.JSON(http.StatusConflict, gin.H{"error": "already-used:email"})
			return
		}

		user, err := mod.RegisterUser(creds.Username, creds.Mail, creds.Password)

		if err != nil {
			fmt.Println("Error registering user:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
			return
		}

		fmt.Println("User registered:", user)
		c.JSON(http.StatusOK, gin.H{"message": "User registered"})
	})

	r.POST("/login", func(c *gin.Context) {
		var creds controller.Credentials
		if err := c.ShouldBindJSON(&creds); err != nil {
			fmt.Println(creds)
			fmt.Println(c.Params.Get("mail"))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		userInfo, err := mod.GetUserByEmailOrUsername(creds.MailOrUsername, false)

		fmt.Println(err)

		isUserVerified := mod.IsUserVerified(userInfo.Email)
		if !isUserVerified {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User is not verified"})
			return
		}

		user, err := mod.AuthenticateUser(creds.MailOrUsername, creds.Password)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		fmt.Println("User authenticated:", user)
		
		if err != nil {
			c.JSON(http.StatusInternalServerError,  gin.H{"error": "Error getting the user"})
			return
		}

		fmt.Println(userInfo)
		tokenString, err := controller.EncodeJWT(userInfo.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return

		}
		c.JSON(http.StatusOK, gin.H{
			"token":    tokenString,
			"username": user.Username,
			"is_admin": user.IsAdmin,
		})
	})

	r.POST("/disconnect", func(c *gin.Context) {
		// Invalidate JWT
		c.JSON(http.StatusOK, gin.H{"message": "Disconnected"})
	})

	r.DELETE("/admin/user/:id", m.AdminAuthenticated(func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		err := mod.RemoveUser(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User removed successfully"})
	}))

	r.DELETE("/user", m.Authenticated(func(c *gin.Context) {
		email, err := controller.GetUserEmailFromGinContext(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
			return
		}

		err = mod.RemoveUser(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User removed successfully"})
	}))

	r.GET("/user",  m.AdminAuthenticated(func (c *gin.Context) {
		users, err := mod.GetAllUser()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}
		c.JSON(http.StatusOK, users)
	}))

	r.POST("/user/verify/:token", func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
			return
		}
		err := mod.VerifyUser(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify user"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "User verified successfully"})
	})

	r.GET("/user/me", m.Authenticated(func(c *gin.Context) {
		userMail, err := controller.GetUserEmailFromGinContext(c)
		fmt.Println("User email from context:", userMail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting the user"})
			return
		}

		user, err := mod.GetUserByEmailOrUsername(userMail, true)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"username": user.Username, "is_admin": user.IsAdmin})
	}))

	r.PUT("/user/change-password", m.Authenticated(func(c *gin.Context) {
		userMail, _ := controller.GetUserEmailFromGinContext(c)
		user, _ := mod.GetUserByEmailOrUsername(userMail,  true)
		var json struct {
			OldPassword string `json:"oldPassword"`
			NewPassword string `json:"newPassword"`
		}

		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if !mod.IsPasswordCorrect(userMail, utils.HashString(json.OldPassword)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "old password is not correct"})
			return
		}

		passwordOk := utils.PasswordValidator(json.NewPassword)

		if !passwordOk {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is not conform"})
			return
		}

		err := mod.ChangeUserPassword(user.ID, json.NewPassword)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error updating password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Password changged successfully"})
	}))

	r.POST("/user/make-admin/:id", m.AdminAuthenticated(func(c *gin.Context) {
		id := c.Param("id")

		err := mod.MakeUserAdmin(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error whihle making the user admin"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "the user is now admin !"})
	}))

	r.GET("/user/orders", m.Authenticated(func(c *gin.Context) {
		email, err := controller.GetUserEmailFromGinContext(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		user, err := mod.GetUserByEmailOrUsername(email, true)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to get user"})
			return
		}

		orders, err := mod.GetAllOrdersFromUser(user.ID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to get orders"})
			return
		}

		c.JSON(200, gin.H{"orders": orders})
	}))

	r.POST("/cart/add/", m.Authenticated(func(c *gin.Context) {
		var prodQuant struct {
			ProductID string `json:"product_id" binding:"required"`
			Quantity  int    `json:"quantity" binding:"required,min=1"`
		}

		if err := c.ShouldBindJSON(&prodQuant); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if prodQuant.ProductID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}
		if prodQuant.Quantity < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be at least 1"})
			return
		}

		userID, err := controller.GetUserEmailFromGinContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		err = mod.AddProductToCart(userID, prodQuant.ProductID, prodQuant.Quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product added to cart successfully"})
	}))

	r.POST("/order", m.Authenticated(func(c *gin.Context) {
		userID, err := controller.GetUserEmailFromGinContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		err = mod.OrderCart(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order created successfully"})
	}))
}

func initProductRoutes(r *gin.Engine) {
	r.GET("/product", m.AdminAuthenticated(func(c *gin.Context) {
		products, err := mod.GetProducts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)
	}))

	r.GET("/product/:id", m.Authenticated(func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		fmt.Println()
		product, err := mod.GetProductByID(id)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while getting the product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"product": product})
	}))

	r.POST("/product", m.AdminAuthenticated(func(c *gin.Context) {
		var product mod.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		fmt.Println(product.Flavors)

		productID, err := mod.AddProduct(&product)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "Product added successfully",
			"productID": productID,
		})
	}))

	r.PUT("/product/:id", m.AdminAuthenticated(func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		var product mod.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if err := mod.UpdateProduct(&product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product updated successfully"})
	}))

	r.POST("/product/:id/image", func(c *gin.Context) {
		productID := c.Param("id")

		// Récupère le fichier image
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(400, gin.H{"error": "Image manquante"})
			return
		}

		path := "uploads/" + file.Filename
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.JSON(500, gin.H{"error": "Échec de l'enregistrement de l'image"})
			return
		}

		// Met à jour le chemin en base
		if err := mod.ChangeImagePath(productID, "/"+path); err != nil {
			c.JSON(500, gin.H{"error": "Échec de la mise à jour de l'image"})
			return
		}

		c.JSON(200, gin.H{"message": "Image mise à jour avec succès", "path": "/" + path})
	})

	r.DELETE("/product/:id", m.AdminAuthenticated(func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		if err := mod.DeleteProduct(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}))
}

func initAdminRoutes(r *gin.Engine) {
	r.GET("/admin", m.AdminAuthenticated(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome Admin!"})
	}))
}

func initFAQRoutes(r *gin.Engine) {
	r.GET("/faq", m.Authenticated(func(c *gin.Context) {
		faqs, err := mod.GetFAQs()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FAQs"})
			return
		}
		c.JSON(http.StatusOK, faqs)
	}))

	r.GET("/faq/:id", m.Authenticated(func(c *gin.Context) {
		id := c.Param("id")
		faq, err := mod.GetFAQ(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FAQ"})
		}
		c.JSON(http.StatusOK, faq)
	}))

	r.POST("/faq", m.AdminAuthenticated(func(c *gin.Context) {
		var faq struct {
			Question string `json:"question" binding:"required"`
			Answer   string `json:"answer" binding:"required"`
		}
		if err := c.ShouldBindJSON(&faq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		err := mod.AddFAQ(faq.Question, faq.Answer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add FAQ"})
			return
		}
		fmt.Println("FAQ added:", faq)
		c.JSON(http.StatusOK, gin.H{"message": "FAQ added successfully"})
	}))

	r.PUT("/faq/:id",  m.AdminAuthenticated(func (c *gin.Context) {
		faqID := c.Param("id")
		var faq struct {
			Question string `json:"question" binding:"required"`
			Answer   string `json:"answer" binding:"required"`
		}

		if err := c.ShouldBindJSON(&faq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		err := mod.UpdateFAQ(faqID, faq.Question, faq.Answer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FAQ"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "FAQ updated successfully"})

	}))

	r.DELETE("/faq/:id", m.AdminAuthenticated(func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "FAQ ID is required"})
			return
		}

		err := mod.DeleteFAQ(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete FAQ"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "FAQ deleted successfully"})
	}))
}

func initLogsRoutes(r *gin.Engine) {
	r.DELETE("/log/:id", m.AdminAuthenticated(func(c *gin.Context) {
		id := c.Param("id")

		err := mod.DeleteLogByID(id)
		if err != nil {
			c.JSON(500, gin.H{"error": "Erreur lors de la suppression du log"})
			return
		}

		c.JSON(200, gin.H{"message": "Log supprimé avec succès"})
	}))

	// Récupérer tous les logs
	r.GET("/log", m.AdminAuthenticated(func(c *gin.Context) {
		logs, err := mod.GetAllLogs()
		if err != nil {
			c.JSON(500, gin.H{"error": "Impossible de récupérer les logs"})
			return
		}

		c.JSON(200, gin.H{"logs": logs})
	}))
}
