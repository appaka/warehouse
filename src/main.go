// https://www.youtube.com/watch?v=SonwZ6MF5BE

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"strconv"

	"net/http"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 54320
	user     = "username"
	password = "password"
	dbname   = "appdatabase"
)

type GetStockResponse struct {
	Warehouse string `json:"warehouse"`
	Stock     int    `json:"stock"`
}

var db *sql.DB

func homePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	//json.NewEncoder(w).Encode(123)
}

func DbInsertTransaction(sku string, warehouse string, quantity int, description string) {
	// INSERT transaction
	// TODO: add dates
	transactionSql := "INSERT INTO public.transaction (sku, warehouse, quantity, description) VALUES ($1, $2, $3, $4);"
	_, transactionErr := db.Exec(transactionSql, sku, warehouse, fmt.Sprintf("+%d", quantity), description)
	if transactionErr != nil {
		panic(transactionErr)
	}

}

func DbUpdateStockAdd(sku string, warehouse string, quantity int) {
	sql := `
		INSERT INTO stock (sku, warehouse, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (sku, warehouse)
		DO UPDATE SET quantity = COALESCE(EXCLUDED.quantity, 0) + $3
		`
	_, updateErr := db.Exec(sql, sku, warehouse, quantity)
	if updateErr != nil {
		panic(updateErr)
	}
}

func DbGetStockBySkuWarehouse(sku string, warehouse string) int {
	var totalStock []byte
	sql := "SELECT quantity FROM stock WHERE sku = $1 AND warehouse = $2;"
	r := db.QueryRow(sql, sku, warehouse)
	selectErr := r.Scan(&totalStock)
	if selectErr != nil {
		panic(selectErr)
	}

	iStock, _ := strconv.Atoi(string(totalStock))
	return iStock
}

func DoAddStock(sku string, warehouse string, quantity int, description string) int {
	// BEGIN
	tx, beginErr := db.Begin()
	if beginErr != nil {
		panic(beginErr)
	}

	// SQL TRANSACTIONS
	DbInsertTransaction(sku, warehouse, quantity, description)
	DbUpdateStockAdd(sku, warehouse, quantity)
	newStock := DbGetStockBySkuWarehouse(sku, warehouse)

	// COMMIT
	commitErr := tx.Commit()
	if commitErr != nil {
		panic(commitErr)
	}

	// Return the new stock
	return newStock
}

func apiAddStock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	sku := params["sku"]
	warehouse := params["warehouse"]
	stock := params["stock"]

	var Description struct {
		Text string
	}

	getBodyErr := json.NewDecoder(r.Body).Decode(&Description)
	if getBodyErr != nil {
		panic(getBodyErr)
	}
	var description = Description.Text
	quantity, _ := strconv.Atoi(stock)

	// Storage
	newStock := DoAddStock(sku, warehouse, quantity, description)

	_, _ = w.Write([]byte(strconv.Itoa(newStock)))
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	conn, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		panic(err)
	}

	db = conn

	DoAddStock("123", "25", 99, "initial stock")

	//router := mux.NewRouter().StrictSlash(true)
	//router.HandleFunc("/", homePage).Methods("GET")
	//
	//router.HandleFunc("/add/sku/{sku}/warehouse/{warehouse}/{stock}", apiAddStock).Methods("GET")
	//
	//log.Fatal(http.ListenAndServe(":8000", router))
}
