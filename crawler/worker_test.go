package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// TestWorkerFetch tests the worker's ability to fetch URLs
func TestWorkerFetch(t *testing.T) {
	// Create a test HTTP server that returns mock content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test Page</body></html>"))
	}))
	defer server.Close()

	// Parse the test server URL
	testURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse test URL: %v", err)
	}

	// Create worker
	input := make(chan *url.URL, 1)
	result := make(chan CrawlResult, 1)
	done := make(chan struct{})
	deadLetter := make(chan *url.URL, 1)

	worker := NewWorker(input, result, done, 0, deadLetter)

	// Start worker in goroutine
	go worker.Start()

	// Send URL to worker
	input <- testURL

	// Wait for result
	select {
	case res := <-result:
		// Verify the result
		if res.Url.String() != testURL.String() {
			t.Errorf("Expected URL %s, got %s", testURL, res.Url)
		}
		if res.ContentType != "text/html" {
			t.Errorf("Expected content type text/html, got %s", res.ContentType)
		}
		if string(res.Body) != "<html><body>Test Page</body></html>" {
			t.Errorf("Unexpected body content: %s", string(res.Body))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for worker result")
	}

	// Clean up
	close(input)
	close(done)
}

// TestWorkerErrorHandling tests that worker sends failed URLs to dead letter
func TestWorkerErrorHandling(t *testing.T) {
	// Create a URL that will fail (invalid host)
	badURL, _ := url.Parse("http://this-host-does-not-exist-12345.com")

	input := make(chan *url.URL, 1)
	result := make(chan CrawlResult, 1)
	done := make(chan struct{})
	deadLetter := make(chan *url.URL, 1)

	worker := NewWorker(input, result, done, 0, deadLetter)
	go worker.Start()

	// Send bad URL
	input <- badURL

	// Should receive in dead letter channel
	select {
	case deadURL := <-deadLetter:
		if deadURL.String() != badURL.String() {
			t.Errorf("Expected dead letter URL %s, got %s", badURL, deadURL)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for dead letter")
	}

	close(input)
	close(done)
}

// TestWorkerGracefulShutdown tests that worker exits cleanly when input channel closes
func TestWorkerGracefulShutdown(t *testing.T) {
	input := make(chan *url.URL)
	result := make(chan CrawlResult, 1)
	done := make(chan struct{})
	deadLetter := make(chan *url.URL, 1)

	worker := NewWorker(input, result, done, 0, deadLetter)

	// Start worker
	go worker.Start()

	// Close input channel immediately (simulates shutdown)
	close(input)

	// Wait for result channel to close (worker should close it on exit)
	select {
	case _, ok := <-result:
		if ok {
			t.Error("Result channel should be closed")
		}
		// Success - channel closed as expected
	case <-time.After(2 * time.Second):
		t.Fatal("Worker did not exit cleanly within timeout")
	}
}

// TestWorkerPolitenessDelay tests that worker respects politeness delay
func TestWorkerPolitenessDelay(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	testURL, _ := url.Parse(server.URL)

	input := make(chan *url.URL, 2)
	result := make(chan CrawlResult, 2)
	done := make(chan struct{})
	deadLetter := make(chan *url.URL, 1)

	worker := NewWorker(input, result, done, 0, deadLetter)
	go worker.Start()

	// Send same URL twice
	input <- testURL

	start := time.Now()

	// Wait for first result
	<-result

	// Send second request to same host
	input <- testURL

	// Wait for second result
	<-result

	elapsed := time.Since(start)

	// Should take at least 2 seconds due to politeness delay
	if elapsed < 2*time.Second {
		t.Errorf("Politeness delay not respected. Elapsed time: %v", elapsed)
	}

	close(input)
	close(done)
}

// TestWorkerHTTPStatusCodes tests handling of different HTTP status codes
func TestWorkerHTTPStatusCodes(t *testing.T) {
	tests := []struct {
		name          string
		statusCode    int
		shouldSucceed bool
	}{
		{"200 OK", http.StatusOK, true},
		{"404 Not Found", http.StatusNotFound, false},
		{"500 Server Error", http.StatusInternalServerError, false},
		{"301 Redirect", http.StatusMovedPermanently, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server with specific status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("Response"))
			}))
			defer server.Close()

			testURL, _ := url.Parse(server.URL)
			input := make(chan *url.URL, 1)
			result := make(chan CrawlResult, 1)
			done := make(chan struct{})
			deadLetter := make(chan *url.URL, 1)

			worker := NewWorker(input, result, done, 0, deadLetter)
			go worker.Start()

			input <- testURL

			select {
			case <-result:
				if !tt.shouldSucceed {
					t.Error("Expected error but got result")
				}
			case <-deadLetter:
				if tt.shouldSucceed {
					t.Error("Expected result but got dead letter")
				}
			case <-time.After(3 * time.Second):
				t.Fatal("Timeout")
			}

			close(input)
			close(done)
		})
	}
}
