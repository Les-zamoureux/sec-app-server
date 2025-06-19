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
	m "sec-app-server/middlewares"
	mod "sec-app-server/model"
)

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	origins := os.Getenv("CLIENT_URL")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{origins},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
	}))

	initUserRoutes(r)
	initProductRoutes(r)
	initFAQRoutes(r)

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
			Mail     string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6,max=100,alphanum"`
		}
		if err := c.ShouldBindJSON(&creds); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		exists, err := mod.CheckUserExists(creds.Mail)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
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

		user, err := mod.AuthenticateUser(creds.Mail, creds.Password)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}

		fmt.Println("User authenticated:", user)

		tokenString, err := controller.EncodeJWT(creds.Mail)
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

	r.DELETE("/user/:email", m.AdminAuthenticated(func(c *gin.Context) {
		email := c.Param("email")
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
			return
		}

		err := mod.RemoveUser(email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User removed successfully"})
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

		userID, err := controller.GetUserFromGinContext(c)
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
		userID, err := controller.GetUserFromGinContext(c)
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
	r.GET("/products", m.AdminAuthenticated(func(c *gin.Context) {
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

		products, err := mod.GetProductsByConditions("ID = " + id)
		if err != nil {
			fmt.Println("Error fetching product:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			return
		}
		if products == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}
		c.JSON(http.StatusOK, products[0])
	}))

	r.POST("/product/add", m.AdminAuthenticated(func(c *gin.Context) {
		var product mod.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := mod.AddProduct(&product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product added successfully"})
	}))

	r.PUT("/product/edit/:id", m.AdminAuthenticated(func(c *gin.Context) {
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

	r.DELETE("/product/delete/:id", m.AdminAuthenticated(func(c *gin.Context) {
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

	r.POST("/admin/create-product", m.AdminAuthenticated(func(c *gin.Context) {
		var product mod.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		if err := mod.AddProduct(&product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Product created successfully"})
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
