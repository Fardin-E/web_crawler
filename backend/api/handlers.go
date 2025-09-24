package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/Fardin-E/web_crawler.git/backend/crawler"
	"github.com/Fardin-E/web_crawler.git/backend/parser"
	"github.com/Fardin-E/web_crawler.git/backend/storage"
)

// A struct to match the JSON request body from the frontend.
type StartCrawlRequest struct {
	URL string `json:"url"`
}

// A struct to match the JSON response sent to the frontend.
type CrawlResponse struct {
	Message string `json:"message"`
}

// StartCrawlHandler reads a URL from a POST request and starts the crawler.
func StartCrawlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	var req StartCrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Make sure the URL is valid before proceeding.
	parsedURL, err := url.Parse(req.URL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		http.Error(w, "Invalid URL provided", http.StatusBadRequest)
		return
	}

	initialUrls := []url.URL{*parsedURL}
	contentParsers := []parser.Parser{&parser.HtmlParser{}}
	contentStorage, _ := storage.NewFileStorage("./data")

	c := crawler.NewCrawler(
		initialUrls,
		contentStorage,
		&crawler.Config{
			MaxRedirects:    5,
			RevisitDelay:    time.Hour * 2,
			WorkerCount:     10,
			ExcludePatterns: []string{},
		},
		contentParsers,
	)

	// Start the crawl in a new goroutine so the API call returns immediately.
	go c.Start()

	w.WriteHeader(http.StatusOK)
	response := CrawlResponse{Message: "Crawl started successfully!"}
	json.NewEncoder(w).Encode(response)
}

// GetCrawlStatusHandler returns the current status of the crawler.
func GetCrawlStatusHandler(w http.ResponseWriter, r *http.Request) {
	// For now, this is a placeholder. You'll need to update this to
	// get and return real-time metrics from your crawler.
	w.Header().Set("Content-Type", "application/json")
	status := map[string]string{"status": "running", "urls_crawled": "123"}
	json.NewEncoder(w).Encode(status)
}
