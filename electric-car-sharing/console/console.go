package main

import (
	"bufio"
	"bytes"
	"electric-car-sharing/services/user-service/models" // Import the User model
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)
type Rental struct {
	ID            int    `json:"id"`
	VehicleID     int    `json:"vehicle_id"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Status        string `json:"status"`
	OvertimeHours int    `json:"overtime_hours"`
	Paid          bool   `json:"paid"`
}

type RentalsResponse struct {
	Rentals []Rental `json:"rentals"`
}

	const userServiceURL = "http://localhost:8080/create-user"
	const loginServiceURL = "http://localhost:8080/login" // The login endpoint
	const viewUserDetailsURL = "http://localhost:8080/view-details" // The view user details endpoint
	const viewUserMembershipURL = "http://localhost:8080/view-membership"
	const updateMembershipURL = "http://localhost:8080/update-membership"
	const viewRentalsURL = "http://localhost:8080/view-rentals"

	var currentUserID int // Variable to store the current logged-in user's ID


	func main() {
		for {
			printMenu()
			option := getUserInput("Enter an option: ")

			switch option {
			case "1":
				createNewUser() // Option 1: Create new user
			case "2":
				if currentUserID == 0 { // Check if user is logged in
					login() // Option 2: Login
				} else {
					logout() // Option 2: Logout if already logged in
				}
			case "3":
				if currentUserID == 0 { // Check if user is logged in
					fmt.Println("User not Logged in") // Show Login option if not logged in
					} else {
					viewUserDetails() // Option 3: View user details (only if logged in)
				}
			case "4":
				if currentUserID == 0 { // Check if user is logged in
					fmt.Println("User not Logged in") // Show Login option if not logged in
					} else {
					updateUserDetails() // Option 4. Update user details (only if logged in)
				}
			case "5":
				if currentUserID == 0 { // Check if user is logged in
					fmt.Println("User not Logged in") // Show Login option if not logged in
					} else {
					updateMembership() // Option 5: Update Membership (only if logged in)
				}
			case "6": // New Option
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				updatePassword() // Option 6: Update Password
			}
			case "7":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				viewAllRentals()
			}
			case "8":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				createRental() // Option 8: Create Rental
			}
		case "9":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				cancelRental() // Option 8: Create Rental
			}
		case "10":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				extendRental() // New function to extend rental
			}
		case "11":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				completeRental() // Call the completeRental function
			}
	
		case "12":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				viewInvoices() // Call the viewInvoices function
			}
		case "13":
			if currentUserID == 0 {
				fmt.Println("User not Logged in")
			} else {
				payInvoice()
			}
		
		
			
			case "0":
				fmt.Println("Goodbye!")
				return
			default:
				fmt.Println("Invalid option. Please try again.")
			}
		}
	}
	// Function to print the menu
	func printMenu() {
		fmt.Println("===================")
		fmt.Println("User Management Console")
		fmt.Println("0. Quit")
		fmt.Println("1. Create new user")
		if currentUserID == 0 {
			fmt.Println("2. Login") // Show Login option if not logged in
		} else {
			fmt.Println("2. Logout") // Show Logout option if logged in
		}
		if currentUserID != 0 {
			fmt.Println("3. View User Details")
			fmt.Println("4. Update User Details")
			fmt.Println("5. Update Membership")
			fmt.Println("6. Update Password")
			fmt.Println("7. View All Rentals")
			fmt.Println("8. Create Rental")
			fmt.Println("9. Cancel Rental")
			fmt.Println("10. Extend Rental")
			fmt.Println("11. Complete Rental")
			fmt.Println("12. View invoices")
			fmt.Println("13. Pay invoice")









		}
	}

	// Function to get user input
	func getUserInput(prompt string) string {
		fmt.Print(prompt)
		var input string
		fmt.Scanln(&input)
		return input
	}

	// Function to create a new user
	func createNewUser() {
		name := getUserInput("Enter user name: ")
		email := getUserInput("Enter user email: ")
		password := getUserInput("Enter user password: ")

		// Reuse the User struct from the models package
		newUser := models.User{
			Name:     name,
			Email:    email,
			Password: password,
		}

		userJSON, err := json.Marshal(newUser)
		if err != nil {
			fmt.Println("Error creating JSON payload:", err)
			return
		}

		req, err := http.NewRequest("POST", userServiceURL, bytes.NewBuffer(userJSON))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Set content type explicitly
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error creating user:", err)
			return
		}
		defer resp.Body.Close()

		// Read response body
		// body, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		//     fmt.Println("Error reading response body:", err)
		//     return
		// }

		// fmt.Printf("Response body: %s\n", body)

		if resp.StatusCode == http.StatusCreated {
			fmt.Println("User created successfully!")
		} else {
			fmt.Printf("Error creating user: %s\n", resp.Status)
		}
	}

	// Function to login the user
	func login() {
		email := getUserInput("Enter email: ")
		password := getUserInput("Enter password: ")

		// Prepare the login data
		loginData := map[string]string{
			"email":    email,
			"password": password,
		}

		// Convert login data to JSON
		loginJSON, err := json.Marshal(loginData)
		if err != nil {
			fmt.Println("Error creating JSON payload:", err)
			return
		}

		// Send the login request to the correct URL
		resp, err := http.Post(loginServiceURL, "application/json", bytes.NewBuffer(loginJSON))
		if err != nil {
			fmt.Println("Error logging in:", err)
			return
		}
		defer resp.Body.Close()

		// Read the response body for debugging
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		// Print the response body for debugging
		fmt.Println("Response body:", string(body))

		// Check the response
		if resp.StatusCode == http.StatusOK {
			// Use the already read response body for decoding
			var userResponse struct {
				UserID int `json:"user_id"`
			}

			// Decode the JSON from the already-read body
			err := json.Unmarshal(body, &userResponse)
			if err != nil {
				fmt.Println("Error decoding response:", err)
				return
			}

			// Save the logged-in user's ID
			currentUserID = userResponse.UserID
			fmt.Println("Login successful! User ID:", currentUserID)
		} else {
			fmt.Println("Invalid credentials. Please try again.")
		}
	}

	// Function to logout the user
	func logout() {
		currentUserID = 0
		fmt.Println("Logged out successfully.")
	}
	func viewUserDetails() {
		if currentUserID == 0 {
			fmt.Println("You must be logged in to view your details.")
			return
		}
	
		// Construct the URL for user details
		url := fmt.Sprintf("%s?user_id=%d", viewUserDetailsURL, currentUserID)
	
		// Send the GET request to retrieve user details
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error retrieving user details:", err)
			return
		}
		defer resp.Body.Close()
	
		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading user details response body:", err)
			return
		}
	
		// Parse the JSON response for user details
		var userDetails map[string]interface{}
		err = json.Unmarshal(body, &userDetails)
		if err != nil {
			fmt.Println("Error parsing user details:", err)
			return
		}
	
		// Construct the URL for membership details
		url2 := fmt.Sprintf("%s?user_id=%d", viewUserMembershipURL, currentUserID)
	
		// Send the GET request to retrieve membership details
		resp2, err := http.Get(url2)
		if err != nil {
			fmt.Println("Error retrieving membership details:", err)
			return
		}
		defer resp2.Body.Close()
	
		// Read the response body
		body2, err := ioutil.ReadAll(resp2.Body)
		if err != nil {
			fmt.Println("Error reading membership details response body:", err)
			return
		}
	
		// Parse the JSON response for membership details
		var membershipDetails map[string]interface{}
		err = json.Unmarshal(body2, &membershipDetails)
		if err != nil {
			fmt.Println("Error parsing membership details:", err)
			return
		}
	
		// Display the formatted details
		fmt.Println("User Details:")
		fmt.Printf("  ID: %v\n", userDetails["user_id"])
		fmt.Printf("  Name: %v\n", userDetails["name"])
		fmt.Printf("  Email: %v\n", userDetails["email"])
		fmt.Printf("  Address: %v\n", userDetails["address"])
		fmt.Printf("  Phone Number: %v\n", userDetails["phone_number"])
		fmt.Printf("  Gender: %v\n", userDetails["gender"])
	
		fmt.Println("\nMembership Details:")
		fmt.Printf("  Membership ID: %v\n", membershipDetails["membership_id"])
		fmt.Printf("  Membership Name: %v\n", membershipDetails["membership_name"])
	}
		// Function to update user details
	func updateUserDetails() {
		if currentUserID == 0 {
			fmt.Println("You must be logged in to update details.")
			return
		}
		reader := bufio.NewReader(os.Stdin)

		// Prompt for each detail
		fmt.Println("Enter new details. Leave blank to skip updating a field.")
		
		fmt.Print("New Address: ")
		address, _ := reader.ReadString('\n')
		address = strings.TrimSpace(address)

		fmt.Print("New Phone Number: ")
		phoneNumber, _ := reader.ReadString('\n')
		phoneNumber = strings.TrimSpace(phoneNumber)

		fmt.Print("New Gender (Male/Female/Other): ")
		gender, _ := reader.ReadString('\n')
		gender = strings.TrimSpace(gender)


		// Create a map to hold the update data
		updateData := make(map[string]string)

		// Add non-empty inputs to the updateData map
		if address != "" {
			updateData["address"] = address
		}
		if phoneNumber != "" {
			updateData["phone_number"] = phoneNumber
		}
		if gender != "" {
			updateData["gender"] = gender
		}

		// If no fields were provided, exit
		if len(updateData) == 0 {
			fmt.Println("No details to update.")
			return
		}

		// Convert the updateData map to JSON
		updateJSON, err := json.Marshal(updateData)
		if err != nil {
			fmt.Println("Error creating JSON payload:", err)
			return
		}

		// Construct the URL with the user_id parameter
		url := fmt.Sprintf("http://localhost:8080/update-details?user_id=%d", currentUserID)

		// Send the POST request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(updateJSON))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error updating details:", err)
			return
		}
		defer resp.Body.Close()

		// Handle the response
		if resp.StatusCode == http.StatusOK {
			fmt.Println("User details updated successfully!")
		} else {
			fmt.Printf("Error updating details: %s\n", resp.Status)
		}
	}



	// Function to update Membership Details
	func updateMembership() {
		if currentUserID == 0 {
			fmt.Println("You must be logged in to update membership.")
			return
		}

		// Display membership options
		fmt.Println("Select a new membership:")
		fmt.Println("1. Basic (0% Discount, No VIP Access)")
		fmt.Println("2. Premium (10% Discount, No VIP Access)")
		fmt.Println("3. VIP (20% Discount, VIP Access)")

		// Get user input
		choice := getUserInput("Enter your choice (1, 2, or 3): ")

		// Validate user input
		var membershipID int
		switch choice {
		case "1":
			membershipID = 1
		case "2":
			membershipID = 2
		case "3":
			membershipID = 3
		default:
			fmt.Println("Invalid choice. Please try again.")
			return
		}


		// Prepare the payload
		updateData := map[string]int{
			"membership_id": membershipID, // Convert string to integer
		}

		// Convert the payload to JSON
		updateJSON, err := json.Marshal(updateData)
		if err != nil {
			fmt.Println("Error creating JSON payload:", err)
			return
		}

		// Construct the URL with the user_id parameter
		url := fmt.Sprintf("%s?user_id=%d", updateMembershipURL, currentUserID)

		// Create a new PUT request
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(updateJSON))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		// Set the appropriate header
		req.Header.Set("Content-Type", "application/json")

		// Execute the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error updating membership:", err)
			return
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}

		// Check the response status
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Membership updated successfully!")
		} else {
			fmt.Printf("Error updating membership: %s\n", resp.Status)
			fmt.Println("Response body:", string(body))
		}
	}

	func updatePassword() {
		if currentUserID == 0 {
			fmt.Println("You must be logged in to update password.")
			return
		}
		// Prompt the user for old password
		var oldPassword, newPassword, confirmPassword string
		fmt.Print("Enter old password: ")
		fmt.Scanln(&oldPassword)

		// Prompt the user for new password
		fmt.Print("Enter new password: ")
		fmt.Scanln(&newPassword)

		// Ask to confirm the new password
		fmt.Print("Confirm new password: ")
		fmt.Scanln(&confirmPassword)

		// Check if the new password and confirmation match
		if newPassword != confirmPassword {
			fmt.Println("Passwords do not match. Try again.")
			return
		}

		// Prepare the request body
		requestBody := map[string]string{
			"old_password": oldPassword,
			"new_password": newPassword,
		}
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			fmt.Println("Error creating request body:", err)
			return
		}

		// Send the POST request to update the password
		url := fmt.Sprintf("http://localhost:8080/update-password?user_id=%d", currentUserID)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		// Execute the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		// Read and display the response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		// Output the response from the server
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Password updated successfully.")
		} else {
			fmt.Printf("Failed to update password: %s\n", string(body))
		}


	}

	// Function to view all rentals
	func viewAllRentals() {
		// Construct the URL with the current user ID as a query parameter
		url := fmt.Sprintf("%s?user_id=%d", viewRentalsURL, currentUserID)
	
		// Send the GET request to the view rentals endpoint
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error retrieving rentals:", err)
			return
		}
		defer resp.Body.Close()
	
		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
	
		// Check the response status
		if resp.StatusCode == http.StatusOK {
			// Parse the JSON response
			var rentalsResponse struct {
				Rentals []struct {
					ID            int    `json:"id"`
					StartDate     string `json:"start_date"`
					EndDate       string `json:"end_date"`
					OvertimeHours int    `json:"overtime_hours"`
					Status        string `json:"status"`
					VehicleID     int    `json:"vehicle_id"`
				} `json:"rentals"`
			}
	
			err = json.Unmarshal(body, &rentalsResponse)
			if err != nil {
				fmt.Println("Error parsing rentals:", err)
				return
			}
	
			// Display the rentals in a formatted manner
			fmt.Println("Rentals:")
			for _, rental := range rentalsResponse.Rentals {
				fmt.Printf(
					"  Rental ID: %d\n  Start Date: %s\n  End Date: %s\n  Overtime Hours: %d\n  Status: %s\n  Vehicle ID: %d\n\n",
					rental.ID, rental.StartDate, rental.EndDate, rental.OvertimeHours, rental.Status, rental.VehicleID,
				)
			}
		} else {
			fmt.Printf("Error retrieving rentals: %s\n", resp.Status)
		}
	}
	

	// Function to create a new rental
	func createRental() {
		if currentUserID == 0 {
			fmt.Println("You must be logged in to create a rental.")
			return
		}
		// Step 1: Fetch available vehicles
		url := fmt.Sprintf("http://localhost:8080/vehicles/available?user_id=%d", currentUserID)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("Error fetching available vehicles:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Error: Unable to fetch vehicles. Status: %s\n", resp.Status)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		// Parse the available vehicles
		var vehicles []map[string]interface{}
		if err := json.Unmarshal(body, &vehicles); err != nil {
			fmt.Println("Error parsing vehicle data:", err)
			return
		}

		if len(vehicles) == 0 {
			fmt.Println("No vehicles available for rental.")
			return
		}

		fmt.Println("Available Vehicles:")
		// Print available vehicle details (id, make, model, year, cost per hour, vip access)
		for _, v := range vehicles {
			fmt.Printf("ID: %v, Make: %v, Model: %v, Year: %v, Cost per Hour: $%.2f, VIP Access: %v\n",
				v["id"], v["make"], v["model"], v["year"], v["cost_per_hour"], v["vip_access"])
		}

		// Step 2: Get user choice for vehicle ID
		vehicleIDStr := getUserInput("Enter the Vehicle ID you want to rent: ")
		vehicleID, err := strconv.Atoi(vehicleIDStr)
		if err != nil {
			fmt.Println("Error: Invalid Vehicle ID. Please enter a valid integer.")
			return
	}

		// Step 3: Get rental hours
		hoursStr := getUserInput("Enter the number of hours you want to rent the vehicle for: ")

		// Convert hours to integer
		hours, err := strconv.Atoi(hoursStr)
		if err != nil || hours <= 0 {
			fmt.Println("Invalid input for hours.")
			return
		}
		// Step 4: Fetch estimated cost using POST request with user_id in the query and other data in the body
	estimateURL := fmt.Sprintf("http://localhost:8080/billing/estimate-cost?user_id=%d", currentUserID)

	estimateData := map[string]interface{}{
		"vehicle_id": vehicleID,
		"hours":      hours,
	}

	// Create JSON payload for estimate data
	estimateJSON, err := json.Marshal(estimateData)
	if err != nil {
		fmt.Println("Error creating JSON payload for estimate:", err)
		return
	}
	// Send the POST request
	resp, err = http.Post(estimateURL, "application/json", bytes.NewBuffer(estimateJSON))
	if err != nil {
		fmt.Println("Error fetching cost estimate:", err)
		return
	}
	defer resp.Body.Close()

	// Log the request body for debugging
	estimateDataJSON, err := json.Marshal(estimateData)
	if err != nil {
		fmt.Println("Error marshaling request data:", err)
		return
	}
	fmt.Println(estimateURL)
	fmt.Printf("Request Data: %s\n", string(estimateDataJSON))  // Log the request body

	// Log the response status and body for debugging
	fmt.Printf("Response Status: %s\n", resp.Status)  // Log the status
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return
	}
	fmt.Printf("Response Body: %s\n", string(bodyBytes))  // Log the response body for debugging

	// Check if the response body contains the estimated cost
	var estimateResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &estimateResp); err != nil {
		fmt.Println("Error parsing estimate data:", err)
		return
	}

	// Check if 'estimated_cost' exists and handle the case if it's missing
	estimatedCost, ok := estimateResp["total_cost"].(float64) // Or check type accordingly
	if !ok {
		fmt.Println("Error: 'total_cost' not found in the response or is of the wrong type")
		return
	}

	// Print the estimated cost if found
	fmt.Printf("Estimated Cost: $%.2f\n", estimatedCost)

		// Step 5: Confirm rental
		confirm := getUserInput("Do you want to confirm the rental? (yes/no): ")
		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Rental cancelled.")
			return
		}

	// Step 6: Create the rental
	createURL := fmt.Sprintf("http://localhost:8080/vehicles/create-rental?user_id=%d", currentUserID)
	rentalData := map[string]interface{}{
		"vehicle_id": vehicleID,
		"hours":      hours,
	}

	rentalJSON, err := json.Marshal(rentalData)
	if err != nil {
		fmt.Println("Error creating JSON payload for rental:", err)
		return
	}

	resp, err = http.Post(createURL, "application/json", bytes.NewBuffer(rentalJSON))
	if err != nil {
		fmt.Println("Error creating rental:", err)
		return
	}
	defer resp.Body.Close()

	// Check if the status code indicates an existing ongoing rental
	if resp.StatusCode == http.StatusConflict {
		fmt.Println("Error: User already has an ongoing rental. Please return the current vehicle before renting another.")
		return
	}

	// Handle the status codes
	if resp.StatusCode == http.StatusConflict {
		fmt.Println("Error: User already has an ongoing rental. Please return the current vehicle before renting another.")
		return
	} else if resp.StatusCode == http.StatusOK {
		fmt.Println("Rental created successfully!")
	} else {
		fmt.Printf("Error creating rental. Status: %s\n", resp.Status)
	}

	}

	// // Utility function to safely convert string to integer
	// func atoi(input string) int {
	// 	value, err := strconv.Atoi(input)
	// 	if err != nil {
	// 		fmt.Println("Invalid input, expected an integer.")
	// 		return 0
	// 	}
	// 	return value
	// }
	func cancelRental() {
		if currentUserID == 0 {
			fmt.Println("You must be logged in to cancel a rental.")
			return
		}
	
		// Step 1: Confirm cancellation
		confirm := getUserInput("Do you want to confirm the rental cancellation? (yes/no): ")
		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Cancellation aborted.")
			return
		}
	
		// Step 2: Construct the URL for canceling the rental
		url := fmt.Sprintf("http://localhost:8080/vehicles/cancel-rental?user_id=%d", currentUserID)
	
		// Step 3: Make the POST request using curl
		cmd := exec.Command("curl", "-s", "-X", "POST", url)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("Error while executing curl command:", err)
			return
		}
	
		// Step 4: Process and format the server response
		response := strings.TrimSpace(string(output)) // Clean up any extra spaces or newlines
		if strings.Contains(response, "No active rentals") {
			fmt.Println("No active rentals found for the user.")
		} else if strings.Contains(response, "Rental cancelled successfully") {
			fmt.Println("Rental canceled successfully!")
		} else {
			fmt.Printf("Unexpected response from the server: %s\n", response)
		}
	}
	
	func extendRental() {
		// Step 1: View active rentals
		fmt.Println("Fetching active rentals...")
		viewRentalsURL := fmt.Sprintf("%s?user_id=%d", viewRentalsURL, currentUserID)
		resp, err := http.Get(viewRentalsURL)
		if err != nil {
			fmt.Println("Error viewing active rentals:", err)
			return
		}
		defer resp.Body.Close()
	
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
	
		// Step 2: Parse the response and get the last rental
		var rentalsResponse RentalsResponse
		err = json.Unmarshal(body, &rentalsResponse)
		if err != nil {
			fmt.Println("Error parsing response:", err)
			return
		}
	
		// Check if there are any rentals and display the last one
		if len(rentalsResponse.Rentals) == 0 {
			fmt.Println("No active rentals found.")
			return
		}
	
		lastRental := rentalsResponse.Rentals[len(rentalsResponse.Rentals)-1]
		if lastRental.Status != "active"{
			println("No Active Rentals")
			return
		}
		fmt.Println("Last Active Rental: ")
		fmt.Printf("Vehicle ID: %d\n", lastRental.VehicleID)
		fmt.Printf("Start Date: %s\n", lastRental.StartDate)
		fmt.Printf("End Date: %s\n", lastRental.EndDate)
		fmt.Printf("Status: %s\n", lastRental.Status)
	
		// Step 3: Select hours to extend
		hoursStr := getUserInput("Enter number of hours to extend: ")
		hours, err := strconv.Atoi(hoursStr)
		if err != nil {
			fmt.Println("Invalid number of hours:", err)
			return
		}
	
		// Step 4: Confirm the rental extension
		confirm := getUserInput("Do you want to extend the rental by " + strconv.Itoa(hours) + " hours? (yes/no): ")
		if confirm != "yes" {
			fmt.Println("Rental extension canceled.")
			return
		}
	
		// Step 5: Extend the rental (POST request)
		extendRentalURL := fmt.Sprintf("http://localhost:8080/vehicles/extend-rental?user_id=%d", currentUserID)
		extendRequestBody := fmt.Sprintf("{\"hours\": %d}", hours)
		req, err := http.NewRequest("POST", extendRentalURL, bytes.NewBuffer([]byte(extendRequestBody)))
		if err != nil {
			fmt.Println("Error creating extend rental request:", err)
			return
		}
	
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error extending rental:", err)
			return
		}
		defer resp.Body.Close()
	
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Rental extended successfully!")
		} else {
			fmt.Printf("Error extending rental: %s\n", resp.Status)
		}
	}

	func completeRental() {
		// Step 1: View active rentals
		fmt.Println("Fetching active rentals...")
		viewRentalsURL := fmt.Sprintf("%s?user_id=%d", viewRentalsURL, currentUserID)
		resp, err := http.Get(viewRentalsURL)
		if err != nil {
			fmt.Println("Error viewing active rentals:", err)
			return
		}
		defer resp.Body.Close()
	
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response body:", err)
			return
		}
	
		// Step 2: Parse the response and get the last rental
		var rentalsResponse RentalsResponse
		err = json.Unmarshal(body, &rentalsResponse)
		if err != nil {
			fmt.Println("Error parsing response:", err)
			return
		}
	
		// Check if there are any rentals and display the last one
		if len(rentalsResponse.Rentals) == 0 {
			fmt.Println("No active rentals found.")
			return
		}
	
		// Fetch the last rental in the response
		lastRental := rentalsResponse.Rentals[len(rentalsResponse.Rentals)-1]

		// Check if the rental status is "active"
		if lastRental.Status == "active" {
			fmt.Println("Last Active Rental: ")
			fmt.Printf("Rental ID: %d\n", lastRental.ID)
			fmt.Printf("Vehicle ID: %d\n", lastRental.VehicleID)
			fmt.Printf("Start Date: %s\n", lastRental.StartDate)
			fmt.Printf("End Date: %s\n", lastRental.EndDate)
			fmt.Printf("Status: %s\n", lastRental.Status)
		} else {
			// If last rental is not active, output a message indicating no active rentals
			fmt.Println("User has no active rentals or the last rental is not active.")
			return
		}


		// Step 3: Confirm the completion of the rental
		confirm := getUserInput("Do you want to complete the rental for Vehicle ID " + strconv.Itoa(lastRental.VehicleID) + "? (yes/no): ")
		if confirm != "yes" {
			fmt.Println("Rental completion canceled.")
			return
		}
	
		// Step 4: Complete the rental (POST request)
		completeRentalURL := fmt.Sprintf("http://localhost:8080/vehicles/complete-rental?user_id=%d", currentUserID)
		req, err := http.NewRequest("POST", completeRentalURL, nil)
		if err != nil {
			fmt.Println("Error creating complete rental request:", err)
			return
		}
	
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error completing rental:", err)
			return
		}
		defer resp.Body.Close()
	
		// Step 5: Print out the invoice (assuming the response contains invoice details)
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Rental completed successfully!")
	
			// Read the response body for the invoice details
			invoiceBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading invoice response:", err)
				return
			}
	
			// Assuming the response contains the invoice details
			fmt.Println("Invoice: ")
			fmt.Println(string(invoiceBody))
		} else {
			fmt.Println("Error completing rental:", resp.Status)
		}
	}
	
	func viewInvoices() {
		fmt.Println("Select an option to view invoices:")
		fmt.Println("1. View all invoices")
		fmt.Println("2. View all unpaid invoices")
		
		// Get the user's selection
		option := getUserInput("Enter an option: ")
		
		switch option {
		case "1":
			viewAllInvoices(false) // View all invoices
		case "2":
			viewAllInvoices(true) // View only unpaid invoices
		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
	
	func viewAllInvoices(unpaidOnly bool) {
		// Create the URL with the necessary query parameters
		url := fmt.Sprintf("http://localhost:8080/billing/get-invoices?userid=%d", currentUserID)
		if unpaidOnly {
			url += "&unpaidonly=true" // Append the unpaidonly=true query parameter
		}
		
		// Make the GET request to the API
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal("Error making GET request:", err)
		}
		defer resp.Body.Close()
		
		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response body:", err)
		}
		
		// Check if the request was successful
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Failed to retrieve invoices:", string(body))
			return
		}
		
		// Parse the JSON response
		var invoices []map[string]interface{}
		err = json.Unmarshal(body, &invoices)
		if err != nil {
			log.Fatal("Error parsing JSON response:", err)
		}
		
		// Display the invoices
		if len(invoices) == 0 {
			fmt.Println("No invoices found.")
			return
		}
		
		fmt.Println("Invoices:")
		for _, invoice := range invoices {
			fmt.Printf("Invoice ID: %v\n", invoice["id"])
			fmt.Printf("Rental ID: %v\n", invoice["rental_id"])
			fmt.Printf("Hours: %v\n", invoice["hours"])
			fmt.Printf("Hours Overdue: %v\n", invoice["hours_overdue"])
			fmt.Printf("Final Cost: $%v\n", invoice["final_cost"])
			fmt.Printf("Paid Status: %v\n", invoice["paid_status"])
			fmt.Printf("Created At: %v\n", invoice["created_at"])
			fmt.Println("----------")
		}
	}
	func payInvoice() {
		// Ask for the invoice ID
		invoiceIDStr := getUserInput("Enter the Invoice ID to pay: ")
		invoiceID, err := strconv.Atoi(invoiceIDStr)
		if err != nil || invoiceID <= 0 {
			fmt.Println("Invalid Invoice ID. Please enter a valid number.")
			return
		}
	
		// Prepare the API URL and payload
		url := fmt.Sprintf("http://localhost:8080/billing/pay-invoice?userid=%d", currentUserID)
		payload := map[string]interface{}{
			"invoice_id": invoiceID,
		}
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Fatal("Error creating JSON payload:", err)
		}
	
		// Make the POST request
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Fatal("Error making POST request:", err)
		}
		defer resp.Body.Close()
	
		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading response body:", err)
		}
	
		// Handle the response
		if resp.StatusCode == http.StatusOK {
			fmt.Println("Invoice payment successful!")
		} else {
			fmt.Printf("Failed to pay invoice: %s\n", string(body))
		}
	}
	