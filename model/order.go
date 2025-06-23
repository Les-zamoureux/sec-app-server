package model

import (
	"fmt"
	"sec-app-server/db"
)

type ContainedProduct struct {
	OrderID   string `json:"order_id"`
	Quantity  int    `json:"quantity"`
	ProductID string `json:"product_id"`
}

type Order struct {
	ID string `json:"id"`
	Numero             string  `json:"numero"`
	Price              float64 `json:"price"`
	Date               string  `json:"date"`
	Status             string  `json:"status"`
	DeliveryCoordinate string  `json:"delivery_coordinate"`
	Products           []ContainedProduct
}

func GetAllOrdersFromUser(userID string) ([]Order, error) {
	orders := []Order{}

	sql, err := db.DB.Query("SELECT id, numero, price, date, status, delivery_coordinate FROM orders o JOIN has_ordered a ON a.order_id=o.id WHERE user_id=$1", userID)
	if err != nil {
		fmt.Println("Error fetching orders:", err)
		return nil, err
	}
	defer sql.Close()

	for sql.Next() {
		var order Order
		if err := sql.Scan(&order.ID, &order.Numero, &order.Price,  &order.Date,  &order.Status, &order.DeliveryCoordinate); err != nil {
			return nil, err
		}

		productRows, err := db.DB.Query(`
			SELECT product_id, quantity 
			FROM contains_product
			WHERE order_id = $1
		`, order.ID)
		if err != nil {
			return nil, err
		}
		defer productRows.Close()

		var containedProducts []ContainedProduct
		for productRows.Next() {
			var cp ContainedProduct
			cp.OrderID = order.ID
			if err := productRows.Scan(&cp.ProductID, &cp.Quantity); err != nil {
				return nil, err
			}
			containedProducts = append(containedProducts, cp)
		}

		order.Products = containedProducts

		orders = append(orders, order)
	}

	return orders, nil
}