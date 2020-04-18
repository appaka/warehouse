package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
)

type Database struct {
	DB *sql.DB
}

func (database *Database) Init(dbHost string, dbPort int, dbUser string, dbPassword string, dbName string) {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)
	conn, err := sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	database.DB = conn
}

func (database *Database) Close() {
	err := database.DB.Close()
	if err != nil {
		panic(err)
	}
}

// PUBLIC

func (database *Database) DoUpdateStock(sku string, warehouse string, quantity int, description string) int {
	// BEGIN
	tx, beginErr := database.DB.Begin()
	if beginErr != nil {
		panic(beginErr)
	}

	// SQL TRANSACTIONS
	database.insertTransaction(sku, warehouse, quantity, description)
	database.updateStock(sku, warehouse, quantity)
	newStock := database.getStockBySkuWarehouse(sku, warehouse)

	// COMMIT
	commitErr := tx.Commit()
	if commitErr != nil {
		panic(commitErr)
	}

	// TODO: save new stock into Redis

	// Return the new stock
	return newStock
}

func (database *Database) GetStock(sku string, warehouse string) map[string]int {
	rows := database.getQueryStock(sku, warehouse)
	defer rows.Close()

	data := make(map[string]int)

	for rows.Next() {
		var rQuantity int
		var rWarehouse string

		if err := rows.Scan(&rWarehouse, &rQuantity); err != nil {
			log.Fatal(err)
		}

		data[rWarehouse] = rQuantity
	}

	return data
}

func (database *Database) GetHistory(sku string, warehouse string) map[string]int {
	rows := database.getQueryHistory(sku, warehouse)
	defer rows.Close()

	data := make(map[string]int)

	for rows.Next() {
		var rQuantity int
		var rDate string
		var rDescription string

		if err := rows.Scan(&rDate, &rQuantity, &rDescription); err != nil {
			log.Fatal(err)
		}

		// TODO: add description
		data[rDate] = rQuantity
	}

	return data
}

// PRIVATE

func (database *Database) getQueryHistory(sku string, warehouse string) *sql.Rows {
	sql := "SELECT inserted_at, quantity, description FROM transaction WHERE sku = ?"
	if warehouse == "" {
		rows, _ := database.DB.Query(sql, sku)
		return rows
	} else {
		sql += " AND warehouse = ?"
		rows, _ := database.DB.Query(sql, sku, warehouse)
		return rows
	}
}

func (database *Database) getQueryStock(sku string, warehouse string) *sql.Rows {
	sql := "SELECT warehouse, quantity FROM stock WHERE sku = ?"
	if warehouse == "" {
		rows, _ := database.DB.Query(sql, sku)
		return rows
	} else {
		sql += " AND warehouse = ?"
		rows, _ := database.DB.Query(sql, sku, warehouse)
		return rows
	}
}

func (database *Database) insertTransaction(sku string, warehouse string, quantity int, description string) {
	// INSERT transaction
	transactionSql := "INSERT INTO public.transaction (sku, warehouse, quantity, description) VALUES ($1, $2, $3, $4);"
	_, transactionErr := database.DB.Exec(transactionSql, sku, warehouse, fmt.Sprintf("+%d", quantity), description)
	if transactionErr != nil {
		panic(transactionErr)
	}

}

func (database *Database) updateStock(sku string, warehouse string, quantity int) {
	sql := `
		INSERT INTO stock (sku, warehouse, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (sku, warehouse)
		DO UPDATE SET quantity = COALESCE(stock.quantity, 0) + $3
		`
	_, updateErr := database.DB.Exec(sql, sku, warehouse, quantity)
	if updateErr != nil {
		panic(updateErr)
	}
}

func (database *Database) getStockBySkuWarehouse(sku string, warehouse string) int {
	// TODO: get data from Redis instead of Postgres
	var totalStock []byte
	sql := "SELECT quantity FROM stock WHERE sku = $1 AND warehouse = $2;"
	r := database.DB.QueryRow(sql, sku, warehouse)
	selectErr := r.Scan(&totalStock)
	if selectErr != nil {
		panic(selectErr)
	}

	iStock, _ := strconv.Atoi(string(totalStock))
	return iStock
}
