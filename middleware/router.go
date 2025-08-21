package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type Application struct {
}

// NewApplicationMiddleware creates a new instance of Application
func NewApplicationMiddleware() *Application {
	return &Application{}
}

func (app *Application) EnableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Add all methods your API will use
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")     // Add any custom headers your frontend sends
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests (OPTIONS method)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK) // Respond with 200 OK for preflight
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs details about incoming requests and their responses.
func (app *Application) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// Log request details
		log.Printf("Incoming Request: Method=%s Path=%s From=%s UserAgent=%s", r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())

		// Call the next handler in the chain
		next.ServeHTTP(w, r)

		// Log response details (you'd need a custom ResponseWriter to capture status code)
		// For simplicity, just log total time for now
		log.Printf("Request Completed: Method=%s Path=%s Duration=%v", r.Method, r.URL.Path, time.Since(start))
	})
}

// RecoveryMiddleware gracefully handles panics to prevent server crashes.
func (app *Application) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err) // Log the panic details
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
