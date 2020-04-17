package database

import (
	"database/sql"
	"fmt"
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

func (database *Database) DoAddStock(sku string, warehouse string, quantity int, description string) int {
	// BEGIN
	tx, beginErr := database.DB.Begin()
	if beginErr != nil {
		panic(beginErr)
	}

	// SQL TRANSACTIONS
	database.insertTransaction(sku, warehouse, quantity, description)
	database.updateStockAdd(sku, warehouse, quantity)
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

// PRIVATE

func (database *Database) insertTransaction(sku string, warehouse string, quantity int, description string) {
	// INSERT transaction
	transactionSql := "INSERT INTO public.transaction (sku, warehouse, quantity, description) VALUES ($1, $2, $3, $4);"
	_, transactionErr := database.DB.Exec(transactionSql, sku, warehouse, fmt.Sprintf("+%d", quantity), description)
	if transactionErr != nil {
		panic(transactionErr)
	}

}

func (database *Database) updateStockAdd(sku string, warehouse string, quantity int) {
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
