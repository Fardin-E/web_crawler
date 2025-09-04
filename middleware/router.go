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

// LoggingMiddleware logs details about incoming requests and their responses.
func (app *Application) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture the status code
		// without interfering with the response stream.
		wrappedWriter := &responseWriter{ResponseWriter: w}

		// Call the next handler in the chain
		next.ServeHTTP(wrappedWriter, r)

		// Log the request and response details after the request is complete
		log.Printf(
			"Request Completed: Method=%s Path=%s Status=%d Duration=%v",
			r.Method,
			r.URL.Path,
			wrappedWriter.status,
			time.Since(start),
		)
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

// responseWriter is a custom http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
