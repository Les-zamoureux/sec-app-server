package model

import (
	"encoding/json"
	"fmt"
	"sec-app-server/db"
)

type Product struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Genetics    string     `json:"genetics"`
	Star        bool       `json:"star"`
	Type        string     `json:"type"`
	Stock       int        `json:"stock"`
	Thc_rate    float64    `json:"thc_rate"`
	Cbd_rate    float64    `json:"cbd_rate"`
	Price       float64    `json:"price"`
	Image       string     `json:"image"`
	Description string     `json:"description"`
	Rating      int        `json:"rating"`
	Color       string     `json:"color"`
	Flavors     []Flavor   `json:"flavors"`
	Aspects     []Aspect   `json:"aspects"`
	Effects     []Effet    `json:"effects"`
	IdealFors   []IdealFor `json:"idealfors"`
}

type Flavor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Aspect struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Effet struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IdealFor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func GetProducts() ([]Product, error) {
	var products []Product
	sql, err := db.DB.Query("SELECT * FROM product")
	if err != nil {
		fmt.Println("Error fetching products:", err)
		return nil, err
	}
	defer sql.Close()

	for sql.Next() {
		var product Product
		if err := sql.Scan(&product.ID, &product.Name, &product.Genetics, &product.Star, &product.Type, &product.Stock, &product.Thc_rate, &product.Cbd_rate, &product.Price, &product.Image, &product.Description, &product.Rating, &product.Color); err != nil {
			fmt.Println(err)
			return nil, err
		}

		product.Aspects = []Aspect{}
		sqlAspect, err := db.DB.Query(`
			SELECT id, name
			FROM aspect a JOIN has_aspect h ON a.id = h.aspect_id
			WHERE product_id = $1
		`, product.ID)
		defer sqlAspect.Close()

		if err != nil {
			fmt.Println("sql::", err)
		}

		for sqlAspect.Next() {
			var aspect Aspect
			sqlAspect.Scan(&aspect.ID, &aspect.Name)
			product.Aspects = append(product.Aspects, aspect)
		}

		product.Flavors = []Flavor{}
		sqlFlavor, err := db.DB.Query(`
			SELECT id, name
			FROM flavor a JOIN has_flavor h ON a.id = h.flavor_id
			WHERE product_id = $1
		`, product.ID)
		defer sqlFlavor.Close()

		if err != nil {
			fmt.Println("sql::", err)
		}

		for sqlFlavor.Next() {
			var flavor Flavor
			sqlFlavor.Scan(&flavor.ID, &flavor.Name)
			product.Flavors = append(product.Flavors, flavor)
		}

		product.IdealFors = []IdealFor{}
		sqlIdeal, err := db.DB.Query(`
			SELECT id, name
			FROM ideal_for a JOIN is_ideal_for h ON a.id = h.ideal_for_id
			WHERE product_id = $1
		`, product.ID)
		defer sqlIdeal.Close()

		if err != nil {
			fmt.Println("sql::", err)
		}

		for sqlIdeal.Next() {
			var idealFor IdealFor
			err := sqlIdeal.Scan(&idealFor.ID, &idealFor.Name)
			if err != nil {
				fmt.Println(err)
			}
			product.IdealFors = append(product.IdealFors, idealFor)
		}

		product.Effects = []Effet{}
		sqlEff, err := db.DB.Query(`
			SELECT id, name
			FROM effet a JOIN has_effect h ON a.id = h.effect_id
			WHERE product_id = $1
		`, product.ID)
		defer sqlEff.Close()

		if err != nil {
			fmt.Println("sql::", err)
		}

		for sqlEff.Next() {
			var effect Effet
			sqlEff.Scan(&effect.ID, &effect.Name)
			product.Effects = append(product.Effects, effect)
		}

		products = append(products, product)
	}
	fmt.Println(products)
	return products, nil
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

func AddProduct(product *Product) (int, error) {
	query := "INSERT INTO product (name, genetics, star, type, stock, thc_rate, cbd_rate, price, image, description, rating, color) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"
	resProd, err := db.DB.Exec(query, product.Name, product.Genetics, product.Star, product.Type, product.Stock, product.Thc_rate, product.Cbd_rate, product.Price, product.Image, product.Description, product.Rating, product.Color)
	fmt.Println(err)
	for _, v := range product.Flavors {
		resFlav, err := db.DB.Exec(`
			INSERT INTO flavor (name)
			SELECT $1
		`, v.Name)

		if err != nil {
			fmt.Println(err)
		}

		lastProd, _ := resProd.LastInsertId()
		lastFlav, _ := resFlav.LastInsertId()

		_, err = db.DB.Exec(`
			INSERT INTO has_flavor (product_id, flavor_id)
			VALUES ($1, $2)
		`, lastProd, lastFlav)

		if err != nil {
			fmt.Println(err)
		}
	}

	for _, v := range product.Aspects {
		resAsp, err := db.DB.Exec(`
			INSERT INTO aspect (name)
			SELECT $1
		`, v.Name)

		if err != nil {
			fmt.Println(err)
		}

		lastProd, _ := resProd.LastInsertId()
		lastAsp, _ := resAsp.LastInsertId()

		_, err = db.DB.Exec(`
			INSERT INTO has_aspect (product_id, aspect_id)
			VALUES ($1, $2)
		`, lastProd, lastAsp)

		if err != nil {
			fmt.Println(err)
		}
	}

	for _, v := range product.Effects {
		resEff, err := db.DB.Exec(`
			INSERT INTO effet (name)
			SELECT $1
			WHERE NOT EXISTS (
				SELECT * FROM effet WHERE name=$1
			)
		`, v.Name)

		if err != nil {
			fmt.Println(err)
		}

		lastProd, _ := resProd.LastInsertId()
		lastEff, _ := resEff.LastInsertId()

		_, err = db.DB.Exec(`
			INSERT INTO has_effect (product_id, effect_id)
			VALUES ($1, $2)
		`, lastProd, lastEff)

		if err != nil {
			fmt.Println(err)
		}
	}

	for _, v := range product.IdealFors {
		resIdeal, err := db.DB.Exec(`
			INSERT INTO ideal_for (name)
			SELECT $1
			WHERE NOT EXISTS (
				SELECT * FROM ideal_for WHERE name=$1
			)
		`, v.Name)

		if err != nil {
			fmt.Println(err)
		}

		lastProd, _ := resProd.LastInsertId()
		lastIdeal, _ := resIdeal.LastInsertId()

		_, err = db.DB.Exec(`
			INSERT INTO is_ideal_for (product_id, ideal_for_id)
			VALUES ($1, $2)
		`, lastProd, lastIdeal)

		if err != nil {
			fmt.Println(err)
		}
	}
	idProduct, _ := resProd.LastInsertId()
	return int(idProduct), err
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
	fmt.Println("id", id)
	query := "SELECT * FROM product WHERE id = $1"
	sql := db.DB.QueryRow(query, id)

	var product Product
	err := sql.Scan(&product.ID, &product.Name, &product.Genetics, &product.Star, &product.Type, &product.Stock, &product.Thc_rate, &product.Cbd_rate, &product.Price, &product.Image, &product.Description, &product.Rating, &product.Color)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	product.Aspects = []Aspect{}
	sqlAspect, err := db.DB.Query(`
			SELECT id, name
			FROM aspect a JOIN has_aspect h ON a.id = h.aspect_id
			WHERE product_id = $1
		`, product.ID)
	defer sqlAspect.Close()

	if err != nil {
		fmt.Println("sql::", err)
	}

	for sqlAspect.Next() {
		var aspect Aspect
		sqlAspect.Scan(&aspect.ID, &aspect.Name)
		product.Aspects = append(product.Aspects, aspect)
	}

	product.Flavors = []Flavor{}
	sqlFlavor, err := db.DB.Query(`
			SELECT id, name
			FROM flavor a JOIN has_flavor h ON a.id = h.flavor_id
			WHERE product_id = $1
		`, product.ID)
	defer sqlFlavor.Close()

	if err != nil {
		fmt.Println("sql::", err)
		return nil, err
	}

	for sqlFlavor.Next() {
		var flavor Flavor
		sqlFlavor.Scan(&flavor.ID, &flavor.Name)
		product.Flavors = append(product.Flavors, flavor)
	}

	product.IdealFors = []IdealFor{}
	sqlIdeal, err := db.DB.Query(`
			SELECT id, name
			FROM ideal_for a JOIN is_ideal_for h ON a.id = h.ideal_for_id
			WHERE product_id = $1
		`, product.ID)
	defer sqlIdeal.Close()

	if err != nil {
		fmt.Println("sql::", err)
		return nil, err
	}

	for sqlIdeal.Next() {
		var idealFor IdealFor
		err := sqlIdeal.Scan(&idealFor.ID, &idealFor.Name)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		product.IdealFors = append(product.IdealFors, idealFor)
	}

	product.Effects = []Effet{}
	sqlEff, err := db.DB.Query(`
			SELECT id, name
			FROM effet a JOIN has_effect h ON a.id = h.effect_id
			WHERE product_id = $1
		`, product.ID)
	defer sqlEff.Close()

	if err != nil {
		fmt.Println("sql::", err)
		return nil, err
	}

	for sqlEff.Next() {
		var effect Effet
		sqlEff.Scan(&effect.ID, &effect.Name)
		product.Effects = append(product.Effects, effect)
	}

	return &product, nil
}
