package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"database/sql"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver

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


func main() {

	initDB()
	defer db.Close()
	// Create a new router
	router := mux.NewRouter()

	

	// Define a route for the root endpoint
	router.HandleFunc("/", test).Methods("GET")
	router.HandleFunc("/create-user", user_handlers.CreateUser(db)).Methods("POST")

	// Add a new route for updating user membership
	router.HandleFunc("/update-membership", user_handlers.UpdateMembership(db)).Methods("PUT")
	router.HandleFunc("/view-membership", user_handlers.ViewMembership(db)).Methods("GET")

	//login
	router.HandleFunc("/login", user_handlers.Login(db)).Methods("POST")
	
	//get details
	router.HandleFunc("/view-details", user_handlers.ViewDetails(db)).Methods("GET")
	router.HandleFunc("/update-details", user_handlers.UpdateDetails(db)).Methods("POST")

	router.HandleFunc("/update-password", user_handlers.UpdatePassword(db)).Methods("POST")
	//get user's rental
	router.HandleFunc("/view-rentals", user_handlers.ViewAllRentals(db)).Methods("GET")




	router.HandleFunc("/vehicles/available", vehicle_handlers.FetchAvailableVehicles(db)).Methods("GET")
	router.HandleFunc("/vehicles/create-rental", vehicle_handlers.CreateRental(db)).Methods("POST")
	router.HandleFunc("/vehicles/cancel-rental", vehicle_handlers.CancelRental(db)).Methods("POST")
	router.HandleFunc("/vehicles/complete-rental", vehicle_handlers.CompleteRental(db)).Methods("POST")
	router.HandleFunc("/vehicles/extend-rental", vehicle_handlers.ExtendRental(db)).Methods("POST")

	router.HandleFunc("/billing/estimate-cost", billing_handlers.EstimateCost(db)).Methods("POST")
	router.HandleFunc("/billing/get-invoices", billing_handlers.FetchInvoices(db)).Methods("GET")
	router.HandleFunc("/billing/pay-invoice", billing_handlers.PayInvoice(db)).Methods("POST")








	// Start the server on port 8080
	fmt.Println("Server running on port 8080")
	http.ListenAndServe(":8080", router)
}


//test should return hello world
//http://localhost:8080 


//created a new user successfully
//curl -X POST http://localhost:8080/create-user -H "Content-Type: application/json" -d "{\"name\": \"John Doe\", \"email\": \"john@example.com\", \"password\": \"securepassword\"}"

//check a user's membership stats
//curl -X GET "http://localhost:8080/view-membership?user_id=1"

//update a user's membership stats
//curl -X PUT http://localhost:8080/update-membership?user_id=1 -H "Content-Type: application/json" -d "{\"membership_id\": 2}"


//test login function
//curl -X POST "http://localhost:8080/login" -H "Content-Type: application/json" -d "{\"email\": \"john@example.com\", \"password\": \"securepassword\"}"


//view details
//curl "http://localhost:8080/view-details?user_id=12"

//update details
//curl -X POST "http://localhost:8080/update-details?user_id=14" -H "Content-Type: application/json" -d "{\"address\": \"123 New St\", \"phone_number\": \"123-456-7890\", \"gender\": \"Female\"}"

//update password
//curl -X POST "http://localhost:8080/update-password?user_id=19" -H "Content-Type: application/json" -d "{\"old_password\":\"password\",\"new_password\":\"password1\"}"

//gets all available vehicles(+ available vip vehicles if user is a vip)
//curl -X GET "http://localhost:8080/vehicles/available?user_id=21" 

//create a new rental(vip cars require user to be vip status)
//curl -X POST "http://localhost:8080/vehicles/create-rental?user_id=16" -H "Content-Type: application/json" -d "{\"vehicle_id\":8,\"hours\":5}"

//cancel a rental
//curl -X POST "http://localhost:8080/vehicles/cancel-rental?user_id=17" 

//complete a rental
//curl -X POST "http://localhost:8080/vehicles/complete-rental?user_id=19" 

//extend a rental
//curl -X POST "http://localhost:8080/vehicles/extend-rental?user_id=16" -H "Content-Type: application/json" -d "{\"hours\": 2}"


//view all user's rental history
//curl -X GET "http://localhost:8080/view-rentals?user_id=16"

//get estimated cost
//curl -X POST "http://localhost:8080/billing/estimate-cost?user_id=21" -H "Content-Type: application/json" -d "{\"vehicle_id\": 9, \"hours\": 5}"

//http://localhost:8080/billing/get-invoices?userid=21&unpaidonly=true	


//curl -X POST "http://localhost:8080/billing/pay-invoice?userid=21" -H "Content-Type: application/json" -d "{\"invoice_id\": 10}"
