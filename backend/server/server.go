package server

import (
	"log"
	"net/http"

	"github.com/Fardin-E/web_crawler.git/backend/api"
	"github.com/rs/cors"
)

// Run starts the API server and sets up all necessary handlers and middleware.
func Run() {
	// Create a new request multiplexer (router).
	mux := http.NewServeMux()

	// Register your API handlers with the router.
	// These handlers should be defined in your backend/api/handlers.go file.
	mux.HandleFunc("/api/status", api.GetCrawlStatusHandler)
	mux.HandleFunc("/api/crawler/start", api.StartCrawlHandler)

	// Add more handlers here as your application grows.

	// Use the CORS middleware to allow requests from your frontend.
	// This is essential for preventing cross-origin security errors.
	handler := cors.Default().Handler(mux)

	// Start the server and bind it to port 8080.
	log.Println("API server listening on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
