package user_handlers

import (
	"database/sql"
	"electric-car-sharing/services/user-service/models"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// CreateUser handles user creation
func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newUser models.User
		err := json.NewDecoder(r.Body).Decode(&newUser)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Validate user details
		if newUser.Name == "" || newUser.Email == "" || newUser.Password == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// Encrypt the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Failed to process password", http.StatusInternalServerError)
			return
		}

		// Insert user into the database
		query := "INSERT INTO users (name, email, password, membership_id) VALUES (?, ?, ?, ?)"
        result, err := db.Exec(query, newUser.Name, newUser.Email, string(hashedPassword), 1) // Default membership ID = 1 (Basic)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

        // Get the last inserted ID
        id, err := result.LastInsertId()
        if err != nil {
            log.Printf("Error retrieving last inserted ID: %v", err)
            http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
            return
        }
		// Set the user ID
		newUser.ID = int(id)

		// Insert corresponding user details with NULL values
		detailsQuery := "INSERT INTO user_details (id, address, phone_number, gender) VALUES (?, NULL, NULL, NULL)"
		_, err = db.Exec(detailsQuery, newUser.ID)
		if err != nil {
			log.Printf("Error inserting user details: %v", err) // Log the error for debugging
			http.Error(w, "Failed to create user details", http.StatusInternalServerError)
			return
		}

        // Prepare the response structure
        response := map[string]interface{}{
            "message": "New user created",
            "user": map[string]interface{}{
                "id":       newUser.ID,
                "name":     newUser.Name,
                "email":    newUser.Email,
                "password": "[PROTECTED]", // Do not return the password in the response
            },
        }

		// Respond with the created user
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

//fetch user details
func ViewDetails(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse and validate the user_id from the query parameters
        userID := r.URL.Query().Get("user_id")
        if userID == "" {
            http.Error(w, "User ID is required", http.StatusBadRequest)
            return
        }

        // Convert userID to an integer
        id, err := strconv.Atoi(userID)
        if err != nil {
            log.Printf("Invalid User ID: %s", userID)
            http.Error(w, "Invalid User ID", http.StatusBadRequest)
            return
        }

        log.Printf("Executing query for user_id: %d", id)

        // Query the database for user details
        // Query the database for user details
        query := `
            SELECT users.id, users.name, users.email, user_details.address, user_details.phone_number, user_details.gender
            FROM users
            LEFT JOIN user_details ON users.id = user_details.id
            WHERE users.id = ?
        `
        log.Printf("Query: %s", query)

        var user struct {
            ID          int    `json:"user_id"`
            Name        string `json:"name"`
            Email       string `json:"email"`
            Address     string `json:"address"`
            PhoneNumber string `json:"phone_number"`
            Gender      string `json:"gender"`
        }

        // Use sql.NullString for nullable fields
        var address, phoneNumber, gender sql.NullString

        err = db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &address, &phoneNumber, &gender)
        if err == sql.ErrNoRows {
            log.Printf("No user found with user_id: %d", id)
            http.Error(w, "User not found", http.StatusNotFound)
            return
        } else if err != nil {
            log.Printf("Error retrieving user details for user_id %d: %v", id, err)
            http.Error(w, "Failed to retrieve user details", http.StatusInternalServerError)
            return
        }

        // Assign nullable fields to user struct, handling NULL values
        user.Address = nullableStringToString(address)
        user.PhoneNumber = nullableStringToString(phoneNumber)
        user.Gender = nullableStringToString(gender)

        log.Printf("Retrieved user details: %+v", user)

        // Respond with the user details
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(user)
    }
}

// Helper function to convert sql.NullString to string
func nullableStringToString(ns sql.NullString) string {
    if ns.Valid {
        return ns.String
    }
    return ""
}

func UpdateDetails(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse and validate the user_id from the query parameters
        userID := r.URL.Query().Get("user_id")
        if userID == "" {
            http.Error(w, "User ID is required", http.StatusBadRequest)
            return
        }

        // Convert userID to an integer
        id, err := strconv.Atoi(userID)
        if err != nil {
            http.Error(w, "Invalid User ID", http.StatusBadRequest)
            return
        }

        // Parse the request body for user details to be updated
        var updateData struct {
            Address     string `json:"address,omitempty"`
            PhoneNumber string `json:"phone_number,omitempty"`
            Gender      string `json:"gender,omitempty"`
        }

        // Decode the JSON request body into the updateData struct
        err = json.NewDecoder(r.Body).Decode(&updateData)
        if err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }

        // Build the SQL query for updating the user's details
        query := `UPDATE user_details SET`
        var args []interface{}
        if updateData.Address != "" {
            query += " address = ?,"
            args = append(args, updateData.Address)
        }
        if updateData.PhoneNumber != "" {
            query += " phone_number = ?,"
            args = append(args, updateData.PhoneNumber)
        }
        if updateData.Gender != "" {
            query += " gender = ?,"
            args = append(args, updateData.Gender)
        }

		// If no fields to update, respond with a message
		if len(args) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "No fields to update",
			})
			return
		}
        // Remove the last comma from the query string
        query = query[:len(query)-1] // Remove the trailing comma
        query += " WHERE id = ?"
        args = append(args, id)

        // Execute the update query
        res, err := db.Exec(query, args...)
        if err != nil {
            http.Error(w, "Failed to update user details", http.StatusInternalServerError)
            return
        }

        // Check if any rows were updated
        rowsAffected, err := res.RowsAffected()
        if err != nil {
            http.Error(w, "Failed to check affected rows", http.StatusInternalServerError)
            return
        }


        if rowsAffected == 0 {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(map[string]string{
                "message": "No changes made; details already up-to-date",
            })
            return
        }

        // Respond with success
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "message": "User details updated successfully",
        })
    }
}


// UpdateMembership handles updating a user's membership
func UpdateMembership(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user_id and membership_id from the request
		type RequestBody struct {
			MembershipID int `json:"membership_id"`
		}

		// Parse query parameters for user ID
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			http.Error(w, "user_id is required", http.StatusBadRequest)
			return
		}

		// Convert user_id to an integer
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}

		// Parse JSON body for membership_id
		var reqBody RequestBody
		err = json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		if reqBody.MembershipID <= 0 {
			http.Error(w, "membership_id must be a positive integer", http.StatusBadRequest)
			return
		}

		// Check the current membership ID for the user
		var currentMembershipID int
		query := "SELECT membership_id FROM users WHERE id = ?"
		err = db.QueryRow(query, userID).Scan(&currentMembershipID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				http.Error(w, "Failed to fetch current membership", http.StatusInternalServerError)
			}
			return
		}

		// If the current membership ID is the same as the new one, return success
		if currentMembershipID == reqBody.MembershipID {
			response := map[string]string{
				"message": "Membership ID is already set to the same value",
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
			return
		}
		// Update the membership in the database
		query = "UPDATE users SET membership_id = ? WHERE id = ?"
		_, err = db.Exec(query, reqBody.MembershipID, userID)
		if err != nil {
			http.Error(w, "Failed to update membership", http.StatusInternalServerError)
			return
		}

		// Respond with success
		response := map[string]string{
			"message": "User membership updated successfully",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		}
}

func ViewMembership(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the user_id from the query parameters
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Query the database for membership details
		query := `
			SELECT users.id, users.name, memberships.id AS membership_id, memberships.name AS membership_name
			FROM users
			LEFT JOIN memberships ON users.membership_id = memberships.id
			WHERE users.id = ?
		`
		var user struct {
			ID             int    `json:"user_id"`
			Name           string `json:"name"`
			MembershipID   int    `json:"membership_id,omitempty"`
			MembershipName string `json:"membership_name,omitempty"`
		}

		err := db.QueryRow(query, userID).Scan(&user.ID, &user.Name, &user.MembershipID, &user.MembershipName)
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Failed to retrieve membership details", http.StatusInternalServerError)
			return
		}

		// Respond with the membership details
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// Login handles user login and returns the user ID if successful
func Login(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginData struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		// Parse the login data
		err := json.NewDecoder(r.Body).Decode(&loginData)
		if err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		// Validate the input
		if loginData.Email == "" || loginData.Password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

        // Query the database for the user's hashed password
        query := "SELECT id, password FROM users WHERE email = ?"
        var userID int
        var hashedPassword string
        err = db.QueryRow(query, loginData.Email).Scan(&userID, &hashedPassword)
        if err == sql.ErrNoRows {
            http.Error(w, "Invalid credentials", http.StatusUnauthorized)
            return
        } else if err != nil {
            http.Error(w, "Error verifying credentials", http.StatusInternalServerError)
            return
        }
		
		// Compare the provided password with the hashed password
		err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(loginData.Password))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Respond with the user ID if login is successful
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"user_id": userID})
	}
}

func UpdatePassword(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request body exists
		if r.Body == nil {
			http.Error(w, "Request body is required", http.StatusBadRequest)
			return
		}

		// Parse the user_id from the query parameters
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert userID to an integer
		id, err := strconv.Atoi(userID)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid User ID", http.StatusBadRequest)
			return
		}

		// Define the request structure
		type RequestBody struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		// Parse the request body
		var reqBody RequestBody
		err = json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate the input
		if reqBody.OldPassword == "" || reqBody.NewPassword == "" {
			http.Error(w, "Old and new passwords are required", http.StatusBadRequest)
			return
		}

		// Fetch the current password for the user
		var currentPasswordHash string
		query := "SELECT password FROM users WHERE id = ?"
		err = db.QueryRow(query, id).Scan(&currentPasswordHash)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "User not found", http.StatusNotFound)
			} else {
				log.Printf("Error fetching password for user %d: %v", id, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		// Compare the old password with the stored password
		err = bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(reqBody.OldPassword))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Hash the new password
		newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(reqBody.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing new password for user %d: %v", id, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Update the password in the database
		updateQuery := "UPDATE users SET password = ? WHERE id = ?"
		_, err = db.Exec(updateQuery, newPasswordHash, id)
		if err != nil {
			log.Printf("Error updating password for user %d: %v", id, err)
			http.Error(w, "Failed to update password", http.StatusInternalServerError)
			return
		}

		// Respond with success
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Password updated successfully",
		})
	}
}

// ViewAllRentals displays all rentals made by a specific user
func ViewAllRentals(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse user ID from query parameters
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

        // Query to get all rentals for the user
        query := `
            SELECT id, vehicle_id, start_date, end_date, status, overtime_hours
            FROM rentals
            WHERE user_id = ?
        `
        rows, err := db.Query(query, userID)
        if err != nil {
            http.Error(w, "Failed to fetch rentals: "+err.Error(), http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        // Create a slice to hold rental records
        var rentals []map[string]interface{}

        // Iterate over the result set
        for rows.Next() {
            var rental struct {
                ID            int
                VehicleID     int
                StartDate     string
                EndDate       string
                Status        string
                OvertimeHours int
            }

            // Scan the row into the rental struct
            err := rows.Scan(&rental.ID, &rental.VehicleID, &rental.StartDate, &rental.EndDate, &rental.Status, &rental.OvertimeHours)
            if err != nil {
                http.Error(w, "Failed to scan rental record: "+err.Error(), http.StatusInternalServerError)
                return
            }

            // Add the rental record to the rentals slice
            rentals = append(rentals, map[string]interface{}{
                "id":             rental.ID,
                "vehicle_id":     rental.VehicleID,
                "start_date":     rental.StartDate,
                "end_date":       rental.EndDate,
                "status":         rental.Status,
                "overtime_hours": rental.OvertimeHours,
            })
        }

        // Check for errors during iteration
        if err := rows.Err(); err != nil {
            http.Error(w, "Error reading rental rows: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Respond with the rental data
        w.Header().Set("Content-Type", "application/json")
        if len(rentals) == 0 {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "message": "No rentals found for the user",
            })
        } else {
            json.NewEncoder(w).Encode(map[string]interface{}{
                "rentals": rentals,
            })
        }
    }
}
