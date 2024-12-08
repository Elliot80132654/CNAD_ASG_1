package handlers

// import (
// 	"database/sql"
// 	"electric-car-sharing/services/vehicle-service/models"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"time"
// )

//GenerateInvoice,rental id,

//EstimateCost, takes in user id vehicle id and hours int

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)


func EstimateCost(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Define struct for the body request
		type EstimateCostRequest struct {
			VehicleID int `json:"vehicle_id"`
			Hours     int `json:"hours"`
		}

		// Define struct for the membership and vehicle
		type Membership struct {
			HourlyRateDiscount int
		}

		type Vehicle struct {
			CostPerHour float64
		}

		// Get UserID from query params
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil || userID <= 0 {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// Decode the JSON body for vehicle and hours
		var req EstimateCostRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Check user membership
		var membership Membership
		query := `
			SELECT m.hourly_rate_discount 
			FROM users u
			JOIN memberships m ON u.membership_id = m.id
			WHERE u.id = ?`
		err = db.QueryRow(query, userID).Scan(&membership.HourlyRateDiscount)
		if err != nil {
			http.Error(w, "User or membership not found", http.StatusNotFound)
			return
		}

		// Get vehicle cost per hour
		var vehicle Vehicle
		query = "SELECT cost_per_hour FROM vehicles WHERE id = ?"
		err = db.QueryRow(query, req.VehicleID).Scan(&vehicle.CostPerHour)
		if err != nil {
			http.Error(w, "Vehicle not found", http.StatusNotFound)
			return
		}

		// Calculate total cost
		discountedRate := vehicle.CostPerHour * (1 - float64(membership.HourlyRateDiscount)/100)
		totalCost := discountedRate * float64(req.Hours)

		// Send response
		response := map[string]interface{}{
			"user_id":    userID,
			"vehicle_id": req.VehicleID,
			"hours":      req.Hours,
			"hourly_rate": discountedRate,
			"total_cost": totalCost,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

type Invoice struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	RentalID    int     `json:"rental_id"`
	Hours       int     `json:"hours"`
	HoursOverdue int    `json:"hours_overdue"`
	FinalCost   float64 `json:"final_cost"`
	PaidStatus  bool    `json:"paid_status"`
	CreatedAt   string  `json:"created_at"`
}

func FetchInvoices(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the URL parameters
		userIDStr := r.URL.Query().Get("userid")
		if userIDStr == "" {
			http.Error(w, "userid is required", http.StatusBadRequest)
			return
		}
		
		// Convert userID from string to int
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid userid", http.StatusBadRequest)
			return
		}

		// Get the unpaidOnly query parameter
		unpaidOnlyStr := r.URL.Query().Get("unpaidonly")
		var unpaidOnly bool
		if unpaidOnlyStr != "" {
			unpaidOnly, err = strconv.ParseBool(unpaidOnlyStr)
			if err != nil {
				http.Error(w, "Invalid unpaidonly value", http.StatusBadRequest)
				return
			}
		}

		// Prepare the query based on unpaidOnly
		var invoices []Invoice
		var query string
		if unpaidOnly {
			query = "SELECT id, user_id, rental_id, hours, hours_overdue, final_cost, paid_status, created_at FROM invoices WHERE user_id = ? AND paid_status = 0"
		} else {
			query = "SELECT id, user_id, rental_id, hours, hours_overdue, final_cost, paid_status, created_at FROM invoices WHERE user_id = ?"
		}

		// Query the database
		rows, err := db.Query(query, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error querying database: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Read the results into the invoices slice
		for rows.Next() {
			var invoice Invoice
			if err := rows.Scan(&invoice.ID, &invoice.UserID, &invoice.RentalID, &invoice.Hours, &invoice.HoursOverdue, &invoice.FinalCost, &invoice.PaidStatus, &invoice.CreatedAt); err != nil {
				http.Error(w, fmt.Sprintf("Error reading rows: %v", err), http.StatusInternalServerError)
				return
			}
			invoices = append(invoices, invoice)
		}

		// Check for errors from iterating over rows
		if err := rows.Err(); err != nil {
			http.Error(w, fmt.Sprintf("Row iteration error: %v", err), http.StatusInternalServerError)
			return
		}

		// Return the invoices as JSON response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(invoices); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
//Pay invoice

func PayInvoice(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the URL parameters
		userIDStr := r.URL.Query().Get("userid")
		if userIDStr == "" {
			http.Error(w, "userid is required", http.StatusBadRequest)
			return
		}

		// Convert userID from string to int
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid userid", http.StatusBadRequest)
			return
		}

		// Parse the invoice ID from the request body
		var requestBody struct {
			InvoiceID int `json:"invoice_id"`
		}

		// Decode the request body into the struct
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Ensure the invoice ID is provided
		if requestBody.InvoiceID == 0 {
			http.Error(w, "invoice_id is required", http.StatusBadRequest)
			return
		}

		// Prepare the update query to set paid_status to true (1)
		query := `UPDATE invoices SET paid_status = 1 WHERE id = ? AND user_id = ?`

		// Execute the query to update the invoice
		result, err := db.Exec(query, requestBody.InvoiceID, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating invoice: %v", err), http.StatusInternalServerError)
			return
		}

		// Check if the invoice was actually updated
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error checking rows affected: %v", err), http.StatusInternalServerError)
			return
		}

		// If no rows were affected, the invoice ID might not exist or doesn't belong to the user
		if rowsAffected == 0 {
			http.Error(w, "Invoice not found or not associated with user", http.StatusNotFound)
			return
		}

		// Respond with a success message
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"message": "Invoice successfully paid",
			"invoice_id": requestBody.InvoiceID,
			"paid_status": true,
		}
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		}
	}
}