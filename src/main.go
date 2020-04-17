// https://www.youtube.com/watch?v=SonwZ6MF5BE

package main

import (
	"./database"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strconv"
)

const VERSION = "0.0.1"

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
	app.showVersion()

	// Database connection
	app.db = database.Database{}
	app.db.Init(dbHost, dbPort, dbUser, dbPassword, dbName)

	// ROUTER
	router := mux.NewRouter().StrictSlash(true)

	//	GET {sku} --> get stock on all warehouses
	router.HandleFunc(baseUri+"/{sku}", app.apiGetStock).Methods("GET")
	//	GET {sku}/{warehouse} --> get stock on that warehouse
	router.HandleFunc(baseUri+"/{sku}/{warehouse}", app.apiGetStock).Methods("GET")

	//	GET _history/{sku} --> get stock history of this product
	router.HandleFunc(baseUri+"/_history/{sku}", app.apiGetHistory).Methods("GET")
	//	GET _history/{sku} --> get stock history of this product and warehouse
	router.HandleFunc(baseUri+"/_history/{sku}/{warehouse}", app.apiGetHistory).Methods("GET")

	//	POST {sku}/{warehouse}/{quantity} --> add stock to this product on this warehouse
	router.HandleFunc(baseUri+"/{sku}/{warehouse}/{quantity}", app.apiAddStock).Methods("POST")

	//	DELETE {sku}/{warehouse}/{quantity} --> subtract stock to this product on this warehouse
	router.HandleFunc(baseUri+"/{sku}/{warehouse}/{quantity}", app.apiSubStock).Methods("DELETE")

	// HTTP server start
	fmt.Printf("listening on :%d\n", portNumber)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portNumber), router))
}

func (app *App) showVersion() {
	fmt.Printf("Appaka Warehouse v%s, by Javier Perez <javier@appaka.ch>, 2020\n", VERSION)
}

// POST http://localhost:8000/api/{sku}/{warehouse}/{quantity}
func (app *App) apiAddStock(w http.ResponseWriter, r *http.Request) {
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

	// save data to database
	newStock := app.db.DoAddStock(sku, warehouse, quantity, "description")

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
func (app *App) apiSubStock(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// GET http://localhost:8000/api/{sku}
// GET http://localhost:8000/api/{sku}/{warehouse}
func (app *App) apiGetStock(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// GET http://localhost:8000/api/_history/{sku}
// GET http://localhost:8000/api/_history/{sku}/{warehouse}
func (app *App) apiGetHistory(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func main() {
	app := App{}
	app.Init()
}
