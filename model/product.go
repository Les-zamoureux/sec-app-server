package model

import (
	"encoding/json"
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
	err := db.DB.QueryRow("SELECT * FROM products").Scan(&products)
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

func GetProductsByConditions(conditions string) ([]Product, error) {
	var products []Product
	query := "SELECT * FROM products WHERE  " + conditions
	err := db.DB.QueryRow(query).Scan(&products)
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
