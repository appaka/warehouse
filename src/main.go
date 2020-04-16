// https://www.youtube.com/watch?v=SonwZ6MF5BE

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 54320
	user     = "username"
	password = "password"
	dbname   = "appdatabase"
)

type AddStockResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Sku       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int    `json:"quantity"`
}

type StockList struct {
	Warehouse string `json:"warehouse"`
	Quantity  int    `json:"quantity"`
}
type GetStockResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Sku     string      `json:"sku"`
	Stock   []StockList `json:"data"`
}

var db *sql.DB

func DbInsertTransaction(sku string, warehouse string, quantity int, description string) {
	// INSERT transaction
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
		DO UPDATE SET quantity = COALESCE(stock.quantity, 0) + $3
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

// POST http://localhost:8000/api/{sku}/{warehouse}/{quantity}
func apiAddStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sku := params["sku"]
	warehouse := params["warehouse"]
	strQuantity := params["quantity"]
	quantity, _ := strconv.Atoi(strQuantity)

	// TODO: get text from r.Body
	//getBodyErr := json.NewDecoder(r.Body).Decode(&Description)
	//if getBodyErr != nil {
	//	//panic(getBodyErr)
	//	log.Fatal(getBodyErr)
	//}
	//var description = Description.Text

	// save data
	newStock := DoAddStock(sku, warehouse, quantity, "description")

	// build response
	response := AddStockResponse{
		Success:   true,
		Message:   fmt.Sprintf("Stock added (%d) to %s on %s! New stock = %d", quantity, sku, warehouse, newStock),
		Sku:       sku,
		Warehouse: warehouse,
		Quantity:  newStock,
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		log.Fatal(encodeErr)
	}
}

// DELETE http://localhost:8000/api/{sku}/{warehouse}/{quantity}
func apiSubStock(w http.ResponseWriter, r *http.Request) {

}

// GET http://localhost:8000/api/{sku}
// GET http://localhost:8000/api/{sku}/{warehouse}
func apiGetStock(w http.ResponseWriter, r *http.Request) {

}

// GET http://localhost:8000/api/_history/{sku}
// GET http://localhost:8000/api/_history/{sku}/{warehouse}
func apiGetHistory(w http.ResponseWriter, r *http.Request) {

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

	router := mux.NewRouter().StrictSlash(true)

	// TODO: get this from env
	var baseUri string = "/api"

	//	GET {sku} --> get stock on all warehouses
	router.HandleFunc(baseUri+"/{sku}", apiGetStock).Methods("GET")
	//	GET {sku}/{warehouse} --> get stock on that warehouse
	router.HandleFunc(baseUri+"/{sku}/{warehouse}", apiGetStock).Methods("GET")

	//	GET _history/{sku} --> get stock history of this product
	router.HandleFunc(baseUri+"/_history/{sku}", apiGetHistory).Methods("GET")
	//	GET _history/{sku} --> get stock history of this product and warehouse
	router.HandleFunc(baseUri+"/_history/{sku}/{warehouse}", apiGetHistory).Methods("GET")

	//	POST {sku}/{warehouse}/{quantity} --> add stock to this product on this warehouse
	router.HandleFunc(baseUri+"/{sku}/{warehouse}/{quantity}", apiAddStock).Methods("POST")

	//	DELETE {sku}/{warehouse}/{quantity} --> subtract stock to this product on this warehouse
	router.HandleFunc(baseUri+"/{sku}/{warehouse}/{quantity}", apiSubStock).Methods("DELETE")

	// TODO: get port number from env
	var portNumber int = 8000
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNumber), router))
}
