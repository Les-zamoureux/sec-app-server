package model

import (
	"fmt"
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

func RegisterUser(username, email, password string) (*User, error) {
	var user User
	sql, err := db.DB.Prepare("INSERT INTO users (username, email, password, is_admin) VALUES ($1, $2, $3, $4)")
	if err != nil {
		fmt.Println("Error registering user:", err)
		return nil, err
	}
	_, err = sql.Exec(username, utils.HashString(email), utils.HashString(password), false)
	if err != nil {
		fmt.Println("Error executing user registration:", err)
		return nil, err
	}
	return &user, nil
}

func RemoveUser(email string) error {
	sql, err := db.DB.Prepare("DELETE FROM users WHERE email = $1")
	if err != nil {
		fmt.Println("Error preparing delete statement:", err)
		return err
	}
	_, err = sql.Exec(utils.HashString(email))
	if err != nil {
		fmt.Println("Error executing delete statement:", err)
		return err
	}
	return nil
}

func CheckUserExists(username, email string) (bool, bool, error) {
	var usernameExists bool
	var emailExists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", utils.HashString(email)).Scan(&emailExists)
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 or username = $2)", username).Scan(&usernameExists)
	if err != nil {
		fmt.Println("Error checking if user exists:", err)
		return false, false, err
	}
	return usernameExists, emailExists, nil
}

func AuthenticateUser(email, password string) (*User, error) {
	var user User
	err := db.DB.QueryRow("SELECT id, username, email, password, is_admin FROM users WHERE email = $1 AND password = $2", utils.HashString(email), utils.HashString(password)).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsAdmin)
	if err != nil {
		fmt.Println("Error authenticating user:", err)
		return nil, err
	}
	return &user, nil
}

func IsUserAdmin(email string) (bool, error) {
	var isAdmin bool
	err := db.DB.QueryRow("SELECT is_admin FROM users WHERE email = $1", utils.HashString(email)).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

func AddProductToCart(userID, productID string, quantity int) error {
	sql, err := db.DB.Prepare("INSERT INTO cart (user_id, product_id, quantity) VALUES ($1, $2, $3)")
	if err != nil {
		fmt.Println("Error preparing add to cart statement:", err)
		return err
	}
	_, err = sql.Exec(userID, productID, quantity)
	if err != nil {
		fmt.Println("Error executing add to cart statement:", err)
		return err
	}
	return nil
}

func OrderCart(userID string) error {
	// get all products in the cart with the quantity
	products := []struct{
		product *Product
		quantity int
	}{}

	sql, err := db.DB.Query("SELECT product_id, quantity FROM cart WHERE user_id = $1", userID)
	if err != nil {
		fmt.Println("Error fetching cart items:", err)
		return err
	}
	defer sql.Close()
	for sql.Next() {
		var productID string
		var quantity int
		if err := sql.Scan(&productID, &quantity); err != nil {
			fmt.Println("Error scanning cart item:", err)
			return err
		}
		
		product, err := GetProductByID(productID)
		if err != nil {
			fmt.Println("Error fetching product:", err)
			return err
		}
		
		products = append(products, struct{
			product *Product
			quantity int
		}{product: product, quantity: quantity})
	}

	// create a new order
	orderID := utils.GenerateRandomString(10) // Generate a random order ID
	sqlOrder, err := db.DB.Prepare("INSERT INTO orders (numero, price, date, status) VALUES ($1, $2, $3, $4)")
	if err != nil {
		fmt.Println("Error preparing order statement:", err)
		return err
	}

	// Calculate total price
	totalPrice := 0.0
	for _, item := range products {
		totalPrice += item.product.Price * float64(item.quantity)
	}
	_, err = sqlOrder.Exec(orderID, fmt.Sprintf("%.2f", totalPrice), utils.GetCurrentDate(), "pending")
	if err != nil {
		fmt.Println("Error executing order statement:", err)
		return err
	}
	// Add products to has_ordered table
	for _, item := range products {
		sqlHasOrdered, err := db.DB.Prepare("INSERT INTO has_ordered (user_id, order_id, quantity) VALUES ($1, $2, $3)")
		if err != nil {
			fmt.Println("Error preparing has_ordered statement:", err)
			return err
		}
		_, err = sqlHasOrdered.Exec(userID, orderID, item.quantity)
		if err != nil {
			fmt.Println("Error executing has_ordered statement:", err)
			return err
		}
	}
	// Clear the cart after ordering
	sqlClearCart, err := db.DB.Prepare("DELETE FROM cart WHERE user_id = $1")
	if err != nil {
		fmt.Println("Error preparing clear cart statement:", err)
		return err
	}
	_, err = sqlClearCart.Exec(userID)
	if err != nil {
		fmt.Println("Error executing clear cart statement:", err)
		return err
	}
	return nil
}
