package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Enable CORS for React frontend
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, x-cf-device-id, x-cf-uid, x-cf-bearer, x-cf-refresh")
}

// Signup handler - accepts any registration for testing
func signupHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var signupReq struct {
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&signupReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Signup: %s %s (%s)", signupReq.FirstName, signupReq.LastName, signupReq.Email)

	// Mock signup - accept any registration
	response := map[string]string{
		"access_token":  "mock-access-token-" + time.Now().Format("20060102150405"),
		"refresh_token": "mock-refresh-token-" + time.Now().Format("20060102150405"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login handler - accepts any email/password for testing
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	// Handle preflight
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loginReq struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Login attempt: %s", loginReq.Email)

	// Mock authentication - accept any email/password
	response := map[string]string{
		"access_token":  "mock-access-token-" + time.Now().Format("20060102150405"),
		"refresh_token": "mock-refresh-token-" + time.Now().Format("20060102150405"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Refresh token handler
func refreshHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Return new access token
	response := map[string]string{
		"access_token": "mock-access-token-refreshed-" + time.Now().Format("20060102150405"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Revoke/logout handler
func revokeHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Token revoked (logout)")

	response := map[string]string{
		"message": "Successfully logged out",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Protected endpoint example - for testing authenticated requests
func profileHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check for bearer token
	bearer := r.Header.Get("x-cf-bearer")
	if bearer == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return mock user profile
	response := map[string]interface{}{
		"email":      r.Header.Get("x-cf-uid"),
		"name":       "Mock User",
		"created_at": "2026-01-01T00:00:00Z",
		"portfolio_value": 10000.50,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Mock stock search endpoint
func searchStockHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	query := r.URL.Query().Get("q")
	
	// Return mock stock data
	stocks := []map[string]interface{}{
		{
			"symbol": "AAPL",
			"name":   "Apple Inc.",
			"price":  175.50,
		},
		{
			"symbol": "GOOGL",
			"name":   "Alphabet Inc.",
			"price":  142.30,
		},
		{
			"symbol": "MSFT",
			"name":   "Microsoft Corporation",
			"price":  378.90,
		},
	}

	response := map[string]interface{}{
		"query":   query,
		"results": stocks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Auth endpoints
	http.HandleFunc("/auth/signup", signupHandler)
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/refresh", refreshHandler)
	http.HandleFunc("/auth/revoke", revokeHandler)

	// Protected endpoints
	http.HandleFunc("/api/profile", profileHandler)
	http.HandleFunc("/api/stocks/search", searchStockHandler)

	port := ":8080"
	log.Printf("🚀 Mock CacheFlow Server running on http://localhost%s", port)
	log.Println("📝 Endpoints:")
	log.Println("   POST /auth/signup")
	log.Println("   POST /auth/login")
	log.Println("   POST /auth/refresh")
	log.Println("   POST /auth/revoke")
	log.Println("   GET  /api/profile")
	log.Println("   GET  /api/stocks/search")
	log.Println("\n✅ Accepts any email/password for login/signup")
	log.Println("✅ CORS enabled for http://localhost:5173")
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}