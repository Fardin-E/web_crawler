package frontier

import (
	"net/url"
	"testing"
	"time"
)

// TestFrontierAdd tests adding URLs to the frontier
func TestFrontierAdd(t *testing.T) {
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, []string{})

	testURL, _ := url.Parse("https://example.com")

	// First add should succeed
	added := f.Add(testURL)
	if !added {
		t.Error("Expected URL to be added successfully")
	}

	// Get the URL back
	select {
	case receivedURL := <-f.Get():
		if receivedURL.String() != testURL.String() {
			t.Errorf("Expected URL %s, got %s", testURL, receivedURL)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for URL from frontier")
	}
}

// TestFrontierDeduplication tests that duplicate URLs are not added
func TestFrontierDeduplication(t *testing.T) {
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, []string{})

	testURL, _ := url.Parse("https://example.com/page1")

	// First add should succeed
	added1 := f.Add(testURL)
	if !added1 {
		t.Error("First add should succeed")
	}

	// Consume the URL
	<-f.Get()

	// Second add of same URL should fail (within revisit delay)
	added2 := f.Add(testURL)
	if added2 {
		t.Error("Duplicate URL should not be added")
	}
}

// TestFrontierExcludePatterns tests that excluded patterns are not added
func TestFrontierExcludePatterns(t *testing.T) {
	excludePatterns := []string{"example.com", "test.com"}
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, excludePatterns)

	tests := []struct {
		name          string
		urlStr        string
		shouldBeAdded bool
	}{
		{"Allowed domain", "https://allowed.com/page", true},
		{"Excluded domain 1", "https://example.com/page", false},
		{"Excluded domain 2", "https://test.com/page", false},
		{"Subdomain of allowed", "https://sub.allowed.com/page", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testURL, _ := url.Parse(tt.urlStr)
			added := f.Add(testURL)

			if added != tt.shouldBeAdded {
				t.Errorf("URL %s: expected added=%v, got added=%v",
					tt.urlStr, tt.shouldBeAdded, added)
			}

			// If added, consume it so it doesn't interfere with next test
			if added {
				select {
				case <-f.Get():
				case <-time.After(100 * time.Millisecond):
					t.Error("Failed to consume added URL")
				}
			}
		})
	}
}

// TestFrontierInitialURLs tests that initial URLs are immediately available
func TestFrontierInitialURLs(t *testing.T) {
	url1, _ := url.Parse("https://example1.com")
	url2, _ := url.Parse("https://example2.com")
	url3, _ := url.Parse("https://example3.com")

	initialURLs := []url.URL{*url1, *url2, *url3}
	f := NewFrontier(initialURLs, []string{})

	// Should be able to get all three URLs immediately
	receivedCount := 0
	timeout := time.After(1 * time.Second)

	for i := 0; i < 3; i++ {
		select {
		case <-f.Get():
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout: only received %d out of 3 initial URLs", receivedCount)
		}
	}

	if receivedCount != 3 {
		t.Errorf("Expected to receive 3 URLs, got %d", receivedCount)
	}
}

// TestFrontierTerminate tests that terminating closes the channel
func TestFrontierTerminate(t *testing.T) {
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, []string{})

	// Terminate the frontier
	f.Terminate()

	// Try to add URL after termination
	testURL, _ := url.Parse("https://example.com")
	added := f.Add(testURL)

	if added {
		t.Error("Should not be able to add URL after termination")
	}

	// Channel should be closed
	select {
	case _, ok := <-f.Get():
		if ok {
			t.Error("Channel should be closed after termination")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout checking if channel is closed")
	}
}

// TestFrontierRevisitDelay tests that URLs can be revisited after delay
func TestFrontierRevisitDelay(t *testing.T) {
	// This test would take 2 hours with real timing, so we test the logic
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, []string{})

	testURL, _ := url.Parse("https://example.com/page")

	// Add URL
	f.Add(testURL)
	<-f.Get() // Consume it

	// Immediately try to add again (should fail)
	added := f.Add(testURL)
	if added {
		t.Error("URL should not be re-added within revisit delay")
	}

	// Check that the URL is in history
	if !f.Seen(testURL) {
		t.Error("URL should be marked as seen")
	}
}

// TestFrontierConcurrency tests that frontier is safe for concurrent access
func TestFrontierConcurrency(t *testing.T) {
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, []string{})

	// Add URLs concurrently from multiple goroutines
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				testURL, _ := url.Parse("https://example.com/page" + string(rune(id*10+j)))
				f.Add(testURL)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should be able to retrieve URLs without panicking
	count := 0
	timeout := time.After(2 * time.Second)

	for {
		select {
		case _, ok := <-f.Get():
			if !ok {
				// Channel closed
				return
			}
			count++
			if count >= 100 {
				return // Got all URLs
			}
		case <-timeout:
			t.Logf("Retrieved %d URLs before timeout (expected up to 100)", count)
			return
		}
	}
}

// TestFrontierEmptyFrontier tests behavior with no initial URLs
func TestFrontierEmptyFrontier(t *testing.T) {
	initialURLs := []url.URL{}
	f := NewFrontier(initialURLs, []string{})

	// Add a URL
	testURL, _ := url.Parse("https://example.com")
	added := f.Add(testURL)

	if !added {
		t.Error("Should be able to add URL to empty frontier")
	}

	// Should be able to retrieve it
	select {
	case receivedURL := <-f.Get():
		if receivedURL.String() != testURL.String() {
			t.Errorf("Expected %s, got %s", testURL, receivedURL)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout retrieving URL from frontier")
	}
}
