package models

// Vehicle represents a vehicle in the system
type Vehicle struct {
	ID          int     `json:"id"`
	Make        string  `json:"make"`
	Model       string  `json:"model"`
	Year        int     `json:"year"`
	Available   bool    `json:"available"`
	VIPAccess   bool    `json:"vip_access"`
	CostPerHour float64 `json:"cost_per_hour"`
}
