// https://www.youtube.com/watch?v=SonwZ6MF5BE

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/appaka/warehouse/database"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const TITLE = "Appaka Warehouse"
const YEAR = 2020
const VERSION = "v0.0.1-alpha.1"
const AUTHOR = "Javier Pérez <javier@appaka.ch>"

type App struct {
	db database.Database
}

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

func getenv(key, fallback string) string {
	if value := os.Getenv(key); len(value) != 0 {
		return value
	}
	return fallback
}

func (app *App) Init() {
	// GET CONFIG FROM ENVIRONMENT
	dbHost := getenv("DB_HOST", "localhost")
	dbPort, _ := strconv.Atoi(getenv("DB_PORT", "54320"))
	dbUser := getenv("DB_USER", "username")
	dbPassword := getenv("DB_PASSWORD", "password")
	dbName := getenv("DB_NAME", "appdatabase")
	baseUri := getenv("BASE_URI", "/api")
	portNumber, _ := strconv.Atoi(getenv("HTTP_PORT", "8000"))

	// Header
	app.showHeader()

	// Database connection
	app.db = database.Database{}
	app.db.Init(dbHost, dbPort, dbUser, dbPassword, dbName)

	// ROUTER
	router := mux.NewRouter().StrictSlash(true)

	//	GET stock/{sku} --> get stock on all warehouses
	router.HandleFunc(baseUri+"/stock/{sku}", app.apiGetStock).Methods("GET")
	//	GET stock/{sku}/{warehouse} --> get stock on that warehouse
	router.HandleFunc(baseUri+"/stock/{sku}/{warehouse}", app.apiGetStock).Methods("GET")

	//	POST stock/{sku}/{warehouse}/{quantity} --> add stock to this product on this warehouse
	router.HandleFunc(baseUri+"/stock/{sku}/{warehouse}/{quantity}", app.apiAddStock).Methods("POST")

	//	DELETE stock/{sku}/{warehouse}/{quantity} --> subtract stock to this product on this warehouse
	router.HandleFunc(baseUri+"/stock/{sku}/{warehouse}/{quantity}", app.apiSubStock).Methods("DELETE")

	//	GET history/{sku} --> get stock history of this product
	router.HandleFunc(baseUri+"/history/{sku}", app.apiGetHistory).Methods("GET")
	//	GET history/{sku}/{warehouse} --> get stock history of this product and warehouse
	router.HandleFunc(baseUri+"/history/{sku}/{warehouse}", app.apiGetHistory).Methods("GET")

	// HTTP server start
	fmt.Printf("listening on :%d\n", portNumber)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNumber), router))
}

func (app *App) showHeader() {
	fmt.Printf("%s %d %s by %s\n", TITLE, YEAR, VERSION, AUTHOR)
}

func (app *App) log(message string) {
	// TODO: write to log file
	fmt.Printf("%s - %s\n", time.Now().Format(time.RFC3339), message)
}

// POST http://localhost:8000/api/stock/{sku}/{warehouse}/{quantity}
func (app *App) apiAddStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	sku := params["sku"]
	warehouse := params["warehouse"]
	strQuantity := params["quantity"]
	quantity, _ := strconv.Atoi(strQuantity)

	// get description from body
	bodyBuffer := new(bytes.Buffer)
	_, _ = bodyBuffer.ReadFrom(r.Body)
	description := bodyBuffer.String()

	app.log(fmt.Sprintf("adding stock (%d) to %s@%s (%s)", quantity, sku, warehouse, description))

	// save data to database
	newStock := app.db.DoAddStock(sku, warehouse, quantity, description)

	app.log(fmt.Sprintf("new stock for %s@%s = %d", sku, warehouse, newStock))

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

// DELETE http://localhost:8000/api/stock/{sku}/{warehouse}/{quantity}
func (app *App) apiSubStock(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// GET http://localhost:8000/api/stock/{sku}
// GET http://localhost:8000/api/stock/{sku}/{warehouse}
func (app *App) apiGetStock(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// GET http://localhost:8000/api/history/{sku}
// GET http://localhost:8000/api/history/{sku}/{warehouse}
func (app *App) apiGetHistory(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func main() {
	app := App{}
	app.Init()

	defer app.db.Close()
}
