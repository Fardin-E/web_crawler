package main

import (
	"bytes" // Import the bytes package
	"encoding/json"
	"fmt"
	"io" // Import the io package
	"net/http"

	"github.com/Fardin-E/web_crawler.git/backend/crawler"
	"github.com/Fardin-E/web_crawler.git/middleware"
	log "github.com/sirupsen/logrus"
)

// The request body for the /api/crawler/start endpoint
type startRequest struct {
	URL string `json:"url"`
}

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/api/crawler/start", handleStartCrawler)

	appMiddleware := middleware.NewApplicationMiddleware()
	wrappedRouter := appMiddleware.RecoveryMiddleware(
		appMiddleware.LoggingMiddleware(
			appMiddleware.EnableCORS(
				router,
			),
		),
	)

	fmt.Println("API server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", wrappedRouter)) // Use log.Fatal to catch ListenAndServe errors
}

// handleStartCrawler is the function that routes the API request to the crawler logic.
func handleStartCrawler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the raw request body FIRST, before anything else attempts to read it.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Internal server error reading request body", http.StatusInternalServerError)
		return
	}

	// IMPORTANT: Re-assign the Body to a new io.ReadCloser because io.ReadAll consumes it.
	// This makes the body available again for json.NewDecoder.
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	log.Printf("Raw Request Body received: %s", string(bodyBytes)) // Log the raw body

	var req startRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding JSON request body: %v", err) // Log the specific JSON decoding error
		http.Error(w, "Bad request: Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		log.Print("Validation Error: URL cannot be empty")
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	go crawler.StartCrawling(req.URL)

	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "Crawler started successfully for " + req.URL}
	json.NewEncoder(w).Encode(response)
}
