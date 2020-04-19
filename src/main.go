// https://www.youtube.com/watch?v=SonwZ6MF5BE

package main

import (
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

type BatchUpdateStockRequest struct {
	Key  string                    `json:"key"`
	Data map[string]map[string]int `json:"data"`
}

type UpdateStockRequest struct {
	Sku         string `json:"sku"`
	Warehouse   string `json:"warehouse"`
	Quantity    int    `json:"quantity"`
	Description string `json:"description"`
}

type BatchUpdateStockResponse struct {
	Success bool                      `json:"success"`
	Key     string                    `json:"key"`
	Data    map[string]map[string]int `json:"data"`
}

type UpdateStockResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Sku       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int    `json:"quantity"`
}

type GetStockRequest struct {
	Sku       string `json:"sku"`
	Warehouse string `json:"warehouse"`
}

type GetStockResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Sku     string         `json:"sku"`
	Data    map[string]int `json:"data"`
}

type GetHistoryRequest struct {
	Sku       string `json:"sku"`
	Warehouse string `json:"warehouse"`
}

type GetHistoryResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Sku     string         `json:"sku"`
	Data    map[string]int `json:"data"`
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
	router.HandleFunc(baseUri+"/stock", app.apiGetStock).Methods("GET")
	router.HandleFunc(baseUri+"/stock", app.apiUpdateStock).Methods("POST")
	router.HandleFunc(baseUri+"/stock/batch", app.apiBatchUpdateStock).Methods("POST")
	router.HandleFunc(baseUri+"/history", app.apiGetHistory).Methods("GET")

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

// POST http://localhost:8000/api/stock/batch
func (app *App) apiBatchUpdateStock(w http.ResponseWriter, r *http.Request) {
	// get json request from body
	request := BatchUpdateStockRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// build response
	response := BatchUpdateStockResponse{
		Success: true,
		Key:     request.Key,
		Data:    make(map[string]map[string]int),
	}

	// save data to database
	// TODO: do it in just one sql transaction?
	app.log(fmt.Sprintf("batch stock update: %s", request.Key))
	for sku, stocks := range request.Data {
		response.Data[sku] = make(map[string]int)
		for warehouse, quantity := range stocks {
			newStock := app.db.DoUpdateStock(sku, warehouse, quantity, request.Key)
			app.log(fmt.Sprintf("new stock for %s@%s = %d", sku, warehouse, newStock))
			response.Data[sku][warehouse] = newStock
		}
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		log.Fatal(encodeErr)
	}
}

// POST http://localhost:8000/api/stock
func (app *App) apiUpdateStock(w http.ResponseWriter, r *http.Request) {
	// get json request from body
	request := UpdateStockRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	app.log(fmt.Sprintf("updating stock (%d) to %s@%s (%s)", request.Quantity, request.Sku, request.Warehouse, request.Description))

	// save data to database
	newStock := app.db.DoUpdateStock(request.Sku, request.Warehouse, request.Quantity, request.Description)

	app.log(fmt.Sprintf("new stock for %s@%s = %d", request.Sku, request.Warehouse, newStock))

	// build response
	response := UpdateStockResponse{
		Success:   true,
		Message:   fmt.Sprintf("Stock updated (%d) to %s@%s = %d", request.Quantity, request.Sku, request.Warehouse, newStock),
		Sku:       request.Sku,
		Warehouse: request.Warehouse,
		Quantity:  newStock,
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		log.Fatal(encodeErr)
	}
}

// GET http://localhost:8000/api/stock
func (app *App) apiGetStock(w http.ResponseWriter, r *http.Request) {
	// get json request from body
	request := GetStockRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	// get stock
	data := app.db.GetStock(request.Sku, request.Warehouse)

	// build response
	response := GetStockResponse{
		Success: true,
		Message: fmt.Sprintf("Stock for %s@%s", request.Sku, request.Warehouse),
		Sku:     request.Sku,
		Data:    data,
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		log.Fatal(encodeErr)
	}
}

// GET http://localhost:8000/api/history
func (app *App) apiGetHistory(w http.ResponseWriter, r *http.Request) {
	// get json request from body
	request := GetHistoryRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Fatal(err)
	}

	data := app.db.GetHistory(request.Sku, request.Warehouse)

	// TODO create proper response
	response := GetHistoryResponse{
		Success: true,
		Message: fmt.Sprintf("History for %s@%s", request.Sku, request.Warehouse),
		Sku:     request.Sku,
		Data:    data,
	}

	// send response
	w.Header().Set("Content-Type", "application/json")
	encodeErr := json.NewEncoder(w).Encode(response)
	if encodeErr != nil {
		log.Fatal(encodeErr)
	}
}

func main() {
	app := App{}
	app.Init()

	defer app.db.Close()
}
