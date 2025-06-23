package model

import (
	"encoding/json"
	"fmt"
	"sec-app-server/db"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Genetics    string  `json:"genetics"`
	Star        bool    `json:"star"`
	Type        string  `json:"type"`
	Thc_rate    float64 `json:"thc_rate"`
	Cbd_rate    float64 `json:"cbd_rate"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Color       string  `json:"color"`
}

func GetProducts() ([]Product, error) {
	var products []Product
	err := db.DB.QueryRow("SELECT * FROM product").Scan(&products)
	if err != nil {
		fmt.Println("Error fetching products:", err)
		return nil, err
	}

	jsonData, err := json.Marshal(products)
	if err != nil {
		return nil, err
	}

	var result []Product
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ChangeImagePath(productID, imagePath string) error {
	_, err := db.DB.Exec("UPDATE product SET image = $1 WHERE id = $2", imagePath, productID)
	return err
}

func GetProductsByConditions(conditions string) ([]Product, error) {
	var products []Product
	// Prepare the SQL query with conditions
	query := fmt.Sprintf("SELECT * FROM product WHERE %s", conditions)
	rows, err := db.DB.Query(query)
	rows.Scan(&products)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(products)
	if err != nil {
		return nil, err
	}

	var result []Product
	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func AddProduct(product *Product) error {
	query := "INSERT INTO product (name, genetics, star, type, thc_rate, cbd_rate, price, description, color) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)"
	_, err := db.DB.Exec(query, product.Name, product.Genetics, product.Star, product.Type, product.Thc_rate, product.Cbd_rate, product.Price, product.Description, product.Color)
	return err
}

func UpdateProduct(product *Product) error {
	query := "UPDATE product SET name = $1, genetics = $2, star = $3, type = $4, thc_rate = $5, cbd_rate = $6, price = $7, description = $8, color = $9 WHERE id = $10"
	_, err := db.DB.Exec(query, product.Name, product.Genetics, product.Star, product.Type, product.Thc_rate, product.Cbd_rate, product.Price, product.Description, product.Color, product.ID)
	return err
}

func DeleteProduct(id string) error {
	query := "DELETE FROM product WHERE id = $1"
	_, err := db.DB.Exec(query, id)
	return err
}

func GetProductByID(id string) (*Product, error) {
	query := "SELECT * FROM product WHERE id = $1"
	row := db.DB.QueryRow(query, id)

	var product Product
	err := row.Scan(&product.ID, &product.Name, &product.Genetics, &product.Star, &product.Type, &product.Thc_rate, &product.Cbd_rate, &product.Price, &product.Description, &product.Color)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("no product found with id %s", id)
		}
		return nil, err
	}

	return &product, nil
}
