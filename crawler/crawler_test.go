package crawler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/Fardin-E/web_crawler.git/storage"
)

// TestCrawlerBasicCrawl tests a basic crawl operation
func TestCrawlerBasicCrawl(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body><a href="/page2">Link</a></body></html>`))
	}))
	defer server.Close()

	// Create temporary storage
	tempDir := t.TempDir()
	contentStorage, err := storage.NewFileStorage(tempDir)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	// Parse server URL
	serverURL, _ := url.Parse(server.URL)
	initialUrls := []url.URL{*serverURL}

	// Create crawler config
	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     2,
		ExcludePatterns: []string{},
	}

	// Create crawler
	crawler := NewCrawler(initialUrls, contentStorage, config)

	// Start crawler in goroutine
	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	// Let it crawl for a bit
	time.Sleep(2 * time.Second)

	// Terminate crawler
	crawler.Terminate()

	// Wait for it to finish
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Fatal("Crawler did not terminate within timeout")
	}

	// Verify that files were created in storage
	files, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	if len(files) == 0 {
		t.Error("Expected crawler to save at least one file")
	}
}

// TestCrawlerGracefulShutdown tests that crawler shuts down cleanly
func TestCrawlerGracefulShutdown(t *testing.T) {
	// Create test server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	serverURL, _ := url.Parse(server.URL)
	initialUrls := []url.URL{*serverURL}

	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     5,
		ExcludePatterns: []string{},
	}

	crawler := NewCrawler(initialUrls, contentStorage, config)

	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	// Immediately terminate
	time.Sleep(100 * time.Millisecond)
	crawler.Terminate()

	// Should exit within reasonable time
	select {
	case <-done:
		// Success - crawler exited cleanly
	case <-time.After(5 * time.Second):
		t.Fatal("Crawler did not shut down gracefully within timeout")
	}
}

// TestCrawlerWithInvalidURL tests handling of invalid URLs
func TestCrawlerWithInvalidURL(t *testing.T) {
	tempDir := t.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	// Create invalid URL
	invalidURL, _ := url.Parse("http://this-does-not-exist-12345.invalid")
	initialUrls := []url.URL{*invalidURL}

	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     2,
		ExcludePatterns: []string{},
	}

	crawler := NewCrawler(initialUrls, contentStorage, config)

	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	// Wait a bit for the invalid URL to be processed and fail
	time.Sleep(3 * time.Second)

	// Terminate since invalid URLs won't naturally end the crawl
	crawler.Terminate()

	// Should complete quickly after termination
	select {
	case <-done:
		// Success - crawler handled invalid URL and terminated
	case <-time.After(5 * time.Second):
		t.Fatal("Crawler did not terminate cleanly after invalid URL")
	}
}

// TestCrawlerExcludePatterns tests that exclude patterns work
func TestCrawlerExcludePatterns(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	serverURL, _ := url.Parse(server.URL)
	initialUrls := []url.URL{*serverURL}

	// Exclude the test server's host
	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     2,
		ExcludePatterns: []string{serverURL.Host},
	}

	crawler := NewCrawler(initialUrls, contentStorage, config)

	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	// Give it a moment to process the exclusion
	time.Sleep(1 * time.Second)

	// Terminate since excluded URLs won't naturally end the crawl
	crawler.Terminate()

	// Should complete quickly after termination
	select {
	case <-done:
		// Success
	case <-time.After(3 * time.Second):
		t.Fatal("Crawler did not terminate cleanly with excluded URL")
	}

	// Should not have crawled anything
	files, _ := os.ReadDir(tempDir)
	if len(files) > 0 {
		t.Error("Expected no files to be saved for excluded URL")
	}
}

// TestCrawlerMultipleWorkers tests that multiple workers process URLs
func TestCrawlerMultipleWorkers(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	// Create multiple URLs
	initialUrls := []url.URL{}
	for i := 0; i < 5; i++ {
		testURL, _ := url.Parse(server.URL + "/page" + string(rune('A'+i)))
		initialUrls = append(initialUrls, *testURL)
	}

	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     3,
		ExcludePatterns: []string{},
	}

	crawler := NewCrawler(initialUrls, contentStorage, config)

	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	// Let it process
	time.Sleep(2 * time.Second)
	crawler.Terminate()

	<-done

	// Should have processed some URLs
	if requestCount == 0 {
		t.Error("Expected at least one request to be processed")
	}
}

// TestCrawlerProcessorExecution tests that custom processors are executed
func TestCrawlerProcessorExecution(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test Content"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	serverURL, _ := url.Parse(server.URL)
	initialUrls := []url.URL{*serverURL}

	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     1,
		ExcludePatterns: []string{},
	}

	crawler := NewCrawler(initialUrls, contentStorage, config)

	// Add custom test processor
	processorCalled := false
	testProcessor := &TestProcessor{
		callback: func(result *CrawlResult) error {
			processorCalled = true
			return nil
		},
	}
	crawler.AddProcessor(testProcessor)

	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	time.Sleep(2 * time.Second)
	crawler.Terminate()
	<-done

	if !processorCalled {
		t.Error("Expected custom processor to be called")
	}
}

// TestProcessor is a helper processor for testing
type TestProcessor struct {
	callback func(*CrawlResult) error
}

func (tp *TestProcessor) Process(result *CrawlResult) error {
	return tp.callback(result)
}

// TestCrawlerEmptyInitialURLs tests crawler with no initial URLs
func TestCrawlerEmptyInitialURLs(t *testing.T) {
	tempDir := t.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	initialUrls := []url.URL{} // Empty

	config := &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour,
		WorkerCount:     2,
		ExcludePatterns: []string{},
	}

	crawler := NewCrawler(initialUrls, contentStorage, config)

	done := make(chan struct{})
	go func() {
		crawler.Start()
		close(done)
	}()

	// Wait a moment for crawler to start
	time.Sleep(500 * time.Millisecond)

	// Terminate since there are no URLs to process
	crawler.Terminate()

	// Should complete quickly (nothing to do)
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Crawler did not terminate with empty URLs")
	}
}

// BenchmarkCrawlerWorkerPool benchmarks the worker pool performance
func BenchmarkCrawlerWorkerPool(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	tempDir := b.TempDir()
	contentStorage, _ := storage.NewFileStorage(tempDir)

	serverURL, _ := url.Parse(server.URL)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		initialUrls := []url.URL{*serverURL}
		config := &Config{
			MaxRedirects:    5,
			RevisitDelay:    time.Hour,
			WorkerCount:     5,
			ExcludePatterns: []string{},
		}

		crawler := NewCrawler(initialUrls, contentStorage, config)

		done := make(chan struct{})
		go func() {
			crawler.Start()
			close(done)
		}()

		time.Sleep(100 * time.Millisecond)
		crawler.Terminate()
		<-done
	}
}
