// https://www.youtube.com/watch?v=SonwZ6MF5BE

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/appaka/warehouse/database"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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
	Sku       string `json:"sku"`
	Warehouse string `json:"warehouse"`
	Quantity  int    `json:"quantity"`
	Key       string `json:"key"`
}

type BatchUpdateStockResponse struct {
	Success bool                      `json:"success"`
	Key     string                    `json:"key"`
	Data    map[string]map[string]int `json:"data"`
}

type UpdateStockResponse struct {
	Success   bool   `json:"success"`
	Key       string `json:"key"`
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
	Sku     string         `json:"sku"`
	Data    map[string]int `json:"data"`
}

type GetHistoryRequest struct {
	Sku       string `json:"sku"`
	Warehouse string `json:"warehouse"`
}

type GetHistoryResponse struct {
	Success bool           `json:"success"`
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
	logPath := getenv("LOG_PATH", "")

	// Logging
	if logPath != "" {
		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		log.SetOutput(file)
	} else {
		log.SetOutput(os.Stdout)
	}

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

	// HTTP server config
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", portNumber),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// HTTP server start
	go func() {
		log.Println("Starting Server")
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	app.waitForShutdown(srv)
}

func (app *App) waitForShutdown(srv *http.Server) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive our signal.
	<-interruptChan

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_ = srv.Shutdown(ctx)

	log.Println("Shutting down")
	os.Exit(0)
}

func (app *App) showHeader() {
	fmt.Printf("%s %d %s by %s\n", TITLE, YEAR, VERSION, AUTHOR)
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
	log.Printf("batch stock update: %s", request.Key)
	for sku, stocks := range request.Data {
		response.Data[sku] = make(map[string]int)
		for warehouse, quantity := range stocks {
			newStock := app.db.DoUpdateStock(sku, warehouse, quantity, request.Key)
			log.Printf("new stock for %s@%s = %d", sku, warehouse, newStock)
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

	// save data to database
	log.Printf("updating stock (%d) to %s@%s (%s)", request.Quantity, request.Sku, request.Warehouse, request.Key)
	newStock := app.db.DoUpdateStock(request.Sku, request.Warehouse, request.Quantity, request.Key)
	log.Printf("new stock for %s@%s = %d", request.Sku, request.Warehouse, newStock)

	// build response
	response := UpdateStockResponse{
		Success:   true,
		Key:       request.Key,
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
