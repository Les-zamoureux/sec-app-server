package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

// InitDB initializes the postgres database connection
func InitDB() error {
	var err error

	DB, err = sql.Open("postgres", "user=user-name password=strong-password dbname=postgres sslmode=disable")
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {	
		return err
	}

	return nil
}

func Test() {
	rows, error := DB.Query("SELECT 1")

	if error != nil {
		panic(error)
	}

	defer rows.Close()

	fmt.Println(rows)
}
