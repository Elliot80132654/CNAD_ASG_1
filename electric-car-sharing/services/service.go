package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
	"github.com/gorilla/mux"

	billing_handlers "electric-car-sharing/services/billing-service/handlers"
	user_handlers "electric-car-sharing/services/user-service/handlers"
	vehicle_handlers "electric-car-sharing/services/vehicle-service/handlers"
)

var db *sql.DB

// Initialize database connection
func initDB() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/electric_car_sharing")
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Database ping error: %v", err)
	}
	fmt.Println("Database connected successfully!")
}
func test(w http.ResponseWriter, r *http.Request) {
	// Respond with "Hello, World!"
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}


func startUserService() {
	router := mux.NewRouter()

	// User service routes
	router.HandleFunc("/", test).Methods("GET")
	router.HandleFunc("/create-user", user_handlers.CreateUser(db)).Methods("POST")
	router.HandleFunc("/update-membership", user_handlers.UpdateMembership(db)).Methods("PUT")
	router.HandleFunc("/view-membership", user_handlers.ViewMembership(db)).Methods("GET")
	router.HandleFunc("/login", user_handlers.Login(db)).Methods("POST")
	router.HandleFunc("/view-details", user_handlers.ViewDetails(db)).Methods("GET")
	router.HandleFunc("/update-details", user_handlers.UpdateDetails(db)).Methods("POST")
	router.HandleFunc("/update-password", user_handlers.UpdatePassword(db)).Methods("POST")
	router.HandleFunc("/view-rentals", user_handlers.ViewAllRentals(db)).Methods("GET")

	// Start server for User service
	fmt.Println("User service running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func startVehicleService() {
	router := mux.NewRouter()

	// Vehicle service routes
	router.HandleFunc("/vehicles/available", vehicle_handlers.FetchAvailableVehicles(db)).Methods("GET")
	router.HandleFunc("/vehicles/create-rental", vehicle_handlers.CreateRental(db)).Methods("POST")
	router.HandleFunc("/vehicles/cancel-rental", vehicle_handlers.CancelRental(db)).Methods("POST")
	router.HandleFunc("/vehicles/complete-rental", vehicle_handlers.CompleteRental(db)).Methods("POST")
	router.HandleFunc("/vehicles/extend-rental", vehicle_handlers.ExtendRental(db)).Methods("POST")

	// Start server for Vehicle service
	fmt.Println("Vehicle service running on port 8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}

func startBillingService() {
	router := mux.NewRouter()

	// Billing service routes
	
	router.HandleFunc("/billing/estimate-cost", billing_handlers.EstimateCost(db)).Methods("POST")
	router.HandleFunc("/billing/get-invoices", billing_handlers.FetchInvoices(db)).Methods("GET")
	router.HandleFunc("/billing/pay-invoice", billing_handlers.PayInvoice(db)).Methods("POST")

	// Start server for Billing service
	fmt.Println("Billing service running on port 8082")
	log.Fatal(http.ListenAndServe(":8082", router))
}

func main() {
	// Initialize database connection
	initDB()
	defer db.Close()

	// Run services in separate goroutines
	go startUserService()
	go startVehicleService()
	go startBillingService()

	// Keep the main thread alive
	select {}
}

// Test should return hello world
// http://localhost:8080
