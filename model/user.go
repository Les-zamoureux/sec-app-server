package model

import (
	"fmt"
	"sec-app-server/db"
	mailcontroller "sec-app-server/mail_controller"
	"sec-app-server/utils"
)

type User struct {
	ID                string    `json:"id"`
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
	token := utils.GenerateRandomString(20) // Generate a random verification token
	sql, err := db.DB.Prepare("INSERT INTO users (username, email, password, is_admin, verification_token, creation_date) VALUES ($1, $2, $3, $4, $5, $6)")
	if err != nil {
		fmt.Println("Error registering user:", err)
		return nil, err
	}
	_, err = sql.Exec(username, utils.HashString(email), utils.HashString(password), false, token, utils.GetCurrentDate())
	if err != nil {
		fmt.Println("Error executing user registration:", err)
		return nil, err
	}
	err = mailcontroller.SendMail(email, "Account Verification", fmt.Sprintf("Please verify your account by clicking the following link: %s/verify-account/%s", utils.ClientUrl, token))

	if err != nil {
		fmt.Println("Error sending verification email:", err)
	}
	return &user, nil
}

func GetUserByEmailOrUsername(emailOrUsername string, alreadyHashed bool) (*User, error) {
	var username string
	var email string
	var isAdmin bool
	var id string
	fmt.Println(emailOrUsername)
	var emaail string
	if !alreadyHashed {
		emaail = utils.HashString(emailOrUsername)
	} else {
		emaail = emailOrUsername
	}

	fmt.Println(emaail, "+", emailOrUsername)
	err := db.DB.QueryRow("SELECT id, username, email, is_admin FROM users WHERE (email=$1 OR username=$2)", emaail, emailOrUsername).Scan(&id, &username, &email, &isAdmin)

	if err != nil {
		fmt.Println("Error fetching user by email:", err)
		return nil, err
	}

	return &User{
		ID: id,
		Email: email,
		Username: username,
		IsAdmin:  isAdmin,
	}, nil
}

func RemoveUserAdmin(id string) error {
	sql, err := db.DB.Prepare("DELETE FROM users WHERE id = $1")
	if err != nil {
		fmt.Println("Error preparing delete statement:", err)
		return err
	}
	_, err = sql.Exec(id)
	if err != nil {
		fmt.Println("Error executing delete statement:", err)
		return err
	}
	return nil
}

func RemoveUser(hashEmail string) error {
	sql, err := db.DB.Prepare("DELETE FROM users WHERE id=$1")
	if err != nil {
		fmt.Println("Error preparing delete statement:", err)
		return err
	}
	_, err = sql.Exec(hashEmail)
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
	err = db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&usernameExists)
	if err != nil {
		fmt.Println("Error checking if user exists:", err)
		return false, false, err
	}
	return usernameExists, emailExists, nil
}

func ChangeUserPassword(id, newPassword string) error {
	sql, err := db.DB.Prepare("UPDATE users SET password=$1 WHERE id=$2")

	if err != nil {
		fmt.Println("Error while preparing the query:", err)
		return err
	}

	_, err = sql.Exec(utils.HashString(newPassword), id)

	if err != nil {
		fmt.Println("Error while changing the password:", err)
		return err
	}

	return nil
}

func GetAllUser() ([]User, error) {
	users := []User{}
	sql, err := db.DB.Query(`
		SELECT id, username, email, password, is_admin, COALESCE(verification_token, ''), COALESCE(verification_date, ''), COALESCE(creation_date, '')
		FROM users
	`)
	if err != nil {
		fmt.Println("Error fetching users:", err)
		return nil, err
	}
	defer sql.Close()

	for sql.Next() {
		var user User
		if err := sql.Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsAdmin, &user.VerificationToken, &user.VerificationDate, &user.CreationDate); err != nil {
			fmt.Println("rrr", err)
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func AuthenticateUser(email, password string) (*User, error) {
	var user User
	err := db.DB.QueryRow("SELECT id, username, email, password, is_admin FROM users WHERE (email = $1 OR username=$2) AND password = $3", utils.HashString(email), email, utils.HashString(password)).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.IsAdmin)
	if err != nil {
		fmt.Println("Error authenticating user:", err)
		return nil, err
	}
	return &user, nil
}

func IsPasswordCorrect(hashEmail, hashPassword string) bool {
	var res bool
	err := db.DB.QueryRow("SELECT EXISTS (SELECT * FROM users WHERE email=$1 AND password=$2)", hashEmail, hashPassword).Scan(&res)
	if err != nil {
		fmt.Println("Error verifying password")
		return false
	}
	return res
}

// func AddToFav(userID, )

func MakeUserAdmin(id string) error {
	sql, err := db.DB.Prepare("UPDATE users SET is_admin = true WHERE id = $1")
	if err != nil {
		fmt.Println("Error preparing verification statement:", err)
		return err
	}

	_, err = sql.Exec(id)
	if err != nil {
		fmt.Println("Error executing update statement:", err)
		return err
	}
	return nil
}


func IsUserAdmin(email string) (bool, error) {
	// only hashed email
	var isAdmin bool
	fmt.Println(email)
	err := db.DB.QueryRow("SELECT is_admin FROM users WHERE email = $1", email).Scan(&isAdmin)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}

func IsUserVerified(email string) bool {
	var isUserVerified bool
	err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT * 
			FROM users 
			WHERE (email=$1 OR username=$1)
			AND verification_token IS NULL 
			AND verification_date IS NOT NULL
		)
	`, email).Scan(&isUserVerified)

	if err != nil {
		return false
	}

	return isUserVerified
}

func VerifyUser(token string) error {
	var userID int
	err := db.DB.QueryRow("SELECT id FROM users WHERE verification_token = $1", token).Scan(&userID)
	if err != nil {
		fmt.Println("Error verifying user:", err)
		return err
	}

	sql, err := db.DB.Prepare("UPDATE users SET verification_token = NULL, verification_date = $1 WHERE id = $2")
	if err != nil {
		fmt.Println("Error preparing verification statement:", err)
		return err
	}
	_, err = sql.Exec(utils.GetCurrentDate(), userID)
	if err != nil {
		fmt.Println("Error executing verification statement:", err)
		return err
	}
	return nil
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
	products := []struct {
		product  *Product
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

		products = append(products, struct {
			product  *Product
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
