package handlers

import (
	"database/sql"
	"electric-car-sharing/services/vehicle-service/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// FetchAvailableVehicles fetches all available vehicles for a given user
func FetchAvailableVehicles(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse user_id from query params
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Check user's membership level
		var vipAccess bool
		query := "SELECT m.vip_access FROM memberships m JOIN users u ON u.membership_id = m.id WHERE u.id = ?"
		err := db.QueryRow(query, userID).Scan(&vipAccess)
		if err == sql.ErrNoRows {
			http.Error(w, "User not found or membership not set", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Failed to fetch membership details", http.StatusInternalServerError)
			return
		}

		// Fetch vehicles based on user's access level
		var vehicleQuery string
		if vipAccess {
			vehicleQuery = "SELECT id, make, model, year, available, vip_access, cost_per_hour FROM vehicles WHERE available = TRUE"
		} else {
			vehicleQuery = "SELECT id, make, model, year, available, vip_access, cost_per_hour FROM vehicles WHERE available = TRUE AND vip_access = FALSE"
		}

		rows, err := db.Query(vehicleQuery)
		if err != nil {
			http.Error(w, "Failed to fetch vehicles", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var vehicles []models.Vehicle
		for rows.Next() {
			var v models.Vehicle
			if err := rows.Scan(&v.ID, &v.Make, &v.Model, &v.Year, &v.Available, &v.VIPAccess, &v.CostPerHour); err != nil {
				http.Error(w, "Failed to parse vehicles", http.StatusInternalServerError)
				return
			}
			vehicles = append(vehicles, v)
		}

		// Respond with available vehicles
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vehicles)
	}
}

// CreateRental creates a new rental and sets the vehicle to unavailable
func CreateRental(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Define the request body structure
		type CreateRentalRequest struct {
			VehicleID int `json:"vehicle_id"`
			Hours     int `json:"hours"`
		}

		// Parse the query parameter for user ID
		userIDParam := r.URL.Query().Get("user_id")
		if userIDParam == "" {
			http.Error(w, "Query parameter 'user_id' is required", http.StatusBadRequest)
			return
		}

		// Convert user ID to an integer
		userID, err := strconv.Atoi(userIDParam)
		if err != nil || userID <= 0 {
			http.Error(w, "Query parameter 'user_id' must be a positive integer", http.StatusBadRequest)
			return
		}

		// Parse the request body
		var reqBody CreateRentalRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate the input
		if reqBody.VehicleID <= 0 || reqBody.Hours <= 0 {
			http.Error(w, "Vehicle ID and Hours must be positive integers", http.StatusBadRequest)
			return
		}

		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		// Check if the user already has an ongoing rental
		var activeRentalExists bool
		activeRentalQuery := "SELECT EXISTS (SELECT 1 FROM rentals WHERE user_id = ? AND status = 'active')"
		err = tx.QueryRow(activeRentalQuery, userID).Scan(&activeRentalExists)
		if err != nil {
			http.Error(w, "Failed to check for existing rentals", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		if activeRentalExists {
			http.Error(w, "User already has an ongoing rental", http.StatusConflict)
			tx.Rollback()
			return
		}


		// Check if the vehicle is available and if it requires VIP access
		var available bool
		var vipOnly bool
		query := "SELECT available, vip_access FROM vehicles WHERE id = ?"
		err = tx.QueryRow(query, reqBody.VehicleID).Scan(&available, &vipOnly)
		if err == sql.ErrNoRows {
			http.Error(w, "Vehicle not found", http.StatusNotFound)
			tx.Rollback()
			return
		} else if err != nil {
			http.Error(w, "Failed to fetch vehicle details", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		if !available {
			http.Error(w, "Vehicle is not available", http.StatusConflict)
			tx.Rollback()
			return
		}

		// If the vehicle requires VIP access, verify user's membership
		if vipOnly {
			var vipAccess bool
			vipQuery := `
				SELECT m.vip_access 
				FROM memberships m 
				JOIN users u ON u.membership_id = m.id 
				WHERE u.id = ?`
			err := tx.QueryRow(vipQuery, userID).Scan(&vipAccess)
			if err == sql.ErrNoRows {
				http.Error(w, "User not found or membership not set", http.StatusNotFound)
				tx.Rollback()
				return
			} else if err != nil {
				http.Error(w, "Failed to fetch membership details", http.StatusInternalServerError)
				tx.Rollback()
				return
			}

			if !vipAccess {
				http.Error(w, "Vehicle requires VIP access, but user is not a VIP", http.StatusForbidden)
				tx.Rollback()
				return
			}
		}

		// Create the rental
		now := time.Now()
		endTime := now.Add(time.Duration(reqBody.Hours) * time.Hour)
		createRentalQuery := `
			INSERT INTO rentals (user_id, vehicle_id, start_date, end_date, status, overtime_hours)
			VALUES (?, ?, ?, ?, 'active', 0)
			`
		_, err = tx.Exec(createRentalQuery, userID, reqBody.VehicleID, now, endTime)
		if err != nil {
			http.Error(w, "Failed to create rental", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		// Set the vehicle to unavailable
		updateVehicleQuery := "UPDATE vehicles SET available = FALSE WHERE id = ?"
		_, err = tx.Exec(updateVehicleQuery, reqBody.VehicleID)
		if err != nil {
			http.Error(w, "Failed to update vehicle availability", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":       "Rental created successfully",
			"user_id":       userID,
			"vehicle_id":    reqBody.VehicleID,
			"start_date":    now.Format(time.RFC3339),
			"end_date":      endTime.Format(time.RFC3339),
			"status":        "active",
			"overtime_hours": 0,
		})
		
	}
}
func CancelRental(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from query parameters
		userIDParam := r.URL.Query().Get("user_id")
		if userIDParam == "" {
			http.Error(w, "Query parameter 'user_id' is required", http.StatusBadRequest)
			return
		}

		// Parse user ID
		var userID int
		_, err := fmt.Sscanf(userIDParam, "%d", &userID)
		if err != nil || userID <= 0 {
			http.Error(w, "Invalid 'user_id' query parameter", http.StatusBadRequest)
			return
		}

		// Fetch active rental for the user
		var rentalID, vehicleID int
		var startDate string // Changed from time.Time to string temporarily for parsing
		query := `
			SELECT id, vehicle_id, start_date
			FROM rentals
			WHERE user_id = ? AND status = 'active'
		`
		err = db.QueryRow(query, userID).Scan(&rentalID, &vehicleID, &startDate)
		

		// Debugging log to check if query is executed correctly
		if err == sql.ErrNoRows {
			http.Error(w, "No active rentals found for the user", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Failed to fetch active rental: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Parse the start_date string into time.Time
		parsedStartDate, err := time.Parse("2006-01-02 15:04:05", startDate)
		if err != nil {
			http.Error(w, "Failed to parse rental start date", http.StatusInternalServerError)
			return
		}
		// Convert start date to Singapore time (UTC +8)
		location, err := time.LoadLocation("Asia/Singapore")
		if err != nil {
			fmt.Println("Failed to load location:", err)
			return
		}
		parsedStartDate = parsedStartDate.In(location)

		// Get current time in Singapore time zone
		currentTime := time.Now().In(location)


		timeDiff := currentTime.Sub(parsedStartDate)
        // Check if the time difference is greater than 1 hour
        if timeDiff > time.Hour {
            fmt.Println("Cancellation is only allowed within 1 hour of the rental start time")
            http.Error(w, "Cancellation is only allowed within 1 hour of the rental start time", http.StatusBadRequest)
            return
        }
		
		
		fmt.Println("Rental start date in Singapore time:", parsedStartDate)
		fmt.Println("Current time in Singapore time:", currentTime)


		//continue the cancellationg if the current time is within an hour
		// Start a transaction
		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
			return
		}

		// Update rental status to 'cancelled'
		cancelRentalQuery := "UPDATE rentals SET status = 'cancelled' WHERE id = ?"
		_, err = tx.Exec(cancelRentalQuery, rentalID)
		if err != nil {
			http.Error(w, "Failed to cancel rental: "+err.Error(), http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		// Set vehicle to available
		updateVehicleQuery := "UPDATE vehicles SET available = TRUE WHERE id = ?"
		_, err = tx.Exec(updateVehicleQuery, vehicleID)
		if err != nil {
			http.Error(w, "Failed to update vehicle availability", http.StatusInternalServerError)
			tx.Rollback()
			return
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
			return
		}

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":   "Rental cancelled successfully",
			"rental_id": rentalID,
			"vehicle_id": vehicleID,
		})
	}
}

// CompleteRental sets the status of a user's active rental to 'completed' and updates the vehicle's availability to true
// and generates an invoice
func CompleteRental(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userIDParam := r.URL.Query().Get("user_id")
        if userIDParam == "" {
            http.Error(w, "Query parameter 'user_id' is required", http.StatusBadRequest)
            return
        }

        userID, err := strconv.Atoi(userIDParam)
        if err != nil || userID <= 0 {
            http.Error(w, "Query parameter 'user_id' must be a positive integer", http.StatusBadRequest)
            return
        }

        tx, err := db.Begin()
        if err != nil {
            http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
            return
        }

        var rentalID, vehicleID int
        var startDate, endDate string
        query := `
        SELECT id, vehicle_id, start_date, end_date
        FROM rentals
        WHERE user_id = ? AND status = 'active'
        `
        err = tx.QueryRow(query, userID).Scan(&rentalID, &vehicleID, &startDate, &endDate)
        if err == sql.ErrNoRows {
            http.Error(w, "No active rentals found for the user", http.StatusNotFound)
            tx.Rollback()
            return
        } else if err != nil {
            http.Error(w, "Failed to fetch active rental: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }

        // Load Singapore timezone
        location, err := time.LoadLocation("Asia/Singapore")
        if err != nil {
            http.Error(w, "Failed to load location", http.StatusInternalServerError)
            tx.Rollback()
            return
        }

        // Parse start and end dates in UTC and convert to Singapore local time
        parsedStartDate, err := time.Parse("2006-01-02 15:04:05", startDate)
        if err != nil {
            http.Error(w, "Failed to parse rental start date", http.StatusInternalServerError)
            tx.Rollback()
            return
        }
        parsedStartDateLocal := parsedStartDate.In(location)

        parsedEndDate, err := time.Parse("2006-01-02 15:04:05", endDate)
        if err != nil {
            http.Error(w, "Failed to parse rental end date", http.StatusInternalServerError)
            tx.Rollback()
            return
        }
        parsedEndDateLocal := parsedEndDate.In(location)

        // Get current local time in Singapore
        currentTimeLocal := time.Now().In(location)

        // Calculate rental hours and overtime (based on local times)
        rentalHours := int(parsedEndDateLocal.Sub(parsedStartDateLocal).Hours())
        if rentalHours < 0 {
            rentalHours = 0
        }

        overtimeHours := 0
        if currentTimeLocal.After(parsedEndDateLocal) {
            overtimeDuration := currentTimeLocal.Sub(parsedEndDateLocal)
            overtimeHours = int(overtimeDuration.Hours())
        }

        var costPerHour float64
        fetchCostQuery := `SELECT cost_per_hour FROM vehicles WHERE id = ?`
        err = tx.QueryRow(fetchCostQuery, vehicleID).Scan(&costPerHour)
        if err != nil {
            http.Error(w, "Failed to fetch vehicle rate: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }
		fmt.Printf("Debug: Vehicle ID %d has an hourly rental rate of %.2f.\n", vehicleID, costPerHour)


        var hourlyRateDiscount float64
        fetchDiscountQuery := `
        SELECT m.hourly_rate_discount
        FROM memberships m
        INNER JOIN users u ON u.membership_id = m.id
        WHERE u.id = ?
        `
        err = tx.QueryRow(fetchDiscountQuery, userID).Scan(&hourlyRateDiscount)
        if err != nil {
            http.Error(w, "Failed to fetch membership discount: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }
        fmt.Printf("Debug: User ID %d has an hourly rate discount of %.2f%%\n", userID, hourlyRateDiscount)
		fmt.Printf("Debug: Rentals hours are: %d \n", rentalHours)
		fmt.Printf("Debug: Overtime hours are: %d \n", overtimeHours)


		// Convert the discount percentage to a decimal
		discountDecimal := hourlyRateDiscount / 100.0

		// Calculate final cost
		overtimeRate := costPerHour * 1.5
		finalCost := (float64(rentalHours) * costPerHour * (1 - discountDecimal)) +
			(float64(overtimeHours) * overtimeRate)

        // Get the current time in UTC for invoice creation
        invoiceTimeUTC := time.Now().UTC()

        // Insert invoice record with UTC time for creation
        invoiceQuery := `
        INSERT INTO invoices (user_id, rental_id, hours, hours_overdue, final_cost, paid_status, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
        `
        _, err = tx.Exec(invoiceQuery, userID, rentalID, rentalHours, overtimeHours, finalCost, false, invoiceTimeUTC)
        if err != nil {
            http.Error(w, "Failed to create invoice: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }

		// Calculate invoice data directly
		invoice := map[string]interface{}{
			"user_id":        userID,
			"rental_id":      rentalID,
			"hours":          rentalHours,
			"hours_overdue":  overtimeHours,
			"final_cost":     finalCost,
			"paid_status":    false, // Can be updated based on payment status
			"created_at":     time.Now().UTC().Format(time.RFC3339), // Use UTC format for timestamp
		}


        // Update rental status and vehicle availability
        completeRentalQuery := "UPDATE rentals SET status = 'completed' WHERE id = ?"
        _, err = tx.Exec(completeRentalQuery, rentalID)
        if err != nil {
            http.Error(w, "Failed to complete rental: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }

        updateVehicleQuery := "UPDATE vehicles SET available = TRUE WHERE id = ?"
        _, err = tx.Exec(updateVehicleQuery, vehicleID)
        if err != nil {
            http.Error(w, "Failed to update vehicle availability", http.StatusInternalServerError)
            tx.Rollback()
            return
        }

        if err := tx.Commit(); err != nil {
            http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message":    "Rental completed successfully",
            "rental_id":  rentalID,
            "vehicle_id": vehicleID,
			"invoice":    invoice, // Include the invoice directly in the response body

        })
    }
}

// ExtendRental extends the rental end date by the number of hours provided in the request
func ExtendRental(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse user ID from query parameters
        userIDParam := r.URL.Query().Get("user_id")
        if userIDParam == "" {
            http.Error(w, "Query parameter 'user_id' is required", http.StatusBadRequest)
            return
        }
        
        // Log to check the received user_id
        fmt.Println("Received user_id:", userIDParam)
        
        // Convert user ID to an integer
        userID, err := strconv.Atoi(userIDParam)
        if err != nil || userID <= 0 {
            http.Error(w, "Query parameter 'user_id' must be a positive integer", http.StatusBadRequest)
            return
        }

        // Parse the number of hours from the request body
        var requestData struct {
            Hours int `json:"hours"`
        }
        err = json.NewDecoder(r.Body).Decode(&requestData)
        if err != nil || requestData.Hours <= 0 {
            http.Error(w, "Invalid or missing 'hours' in request body", http.StatusBadRequest)
            return
        }

        // Start a transaction
        tx, err := db.Begin()
        if err != nil {
            http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
            return
        }

        // Fetch the active rental for the user
        var rentalID int
        var vehicleID int
        var endDateStr string

        query := `
            SELECT id, vehicle_id, end_date
            FROM rentals
            WHERE user_id = ? AND status = 'active'
        `
		err = tx.QueryRow(query, userID).Scan(&rentalID, &vehicleID, &endDateStr)
        if err == sql.ErrNoRows {
            http.Error(w, "No active rentals found for the user", http.StatusNotFound)
            tx.Rollback()
            return
        } else if err != nil {
            http.Error(w, "Failed to fetch active rental: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }
		// Parse the end date string into time.Time
		parsedEndDate, err := time.Parse("2006-01-02 15:04:05", endDateStr)
		if err != nil {
			http.Error(w, "Failed to parse rental end date", http.StatusInternalServerError)
			tx.Rollback()
			return
		}
		// Add the specified number of hours to the end date
		newEndDate := parsedEndDate.Add(time.Duration(requestData.Hours) * time.Hour)
	
        // Update the rental's end date in the database
        updateEndDateQuery := "UPDATE rentals SET end_date = ? WHERE id = ?"
        _, err = tx.Exec(updateEndDateQuery, newEndDate.Format("2006-01-02 15:04:05"), rentalID)
        if err != nil {
            http.Error(w, "Failed to update end date: "+err.Error(), http.StatusInternalServerError)
            tx.Rollback()
            return
        }
        // Commit the transaction
        if err := tx.Commit(); err != nil {
            http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
            return
        }

        // Respond with success
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "message":    "Rental extended successfully",
            "rental_id":  rentalID,
            "vehicle_id": vehicleID,
            "new_end_date": newEndDate.Format("2006-01-02 15:04:05"),
        })
    }
}
