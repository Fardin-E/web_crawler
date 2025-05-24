package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/Fardin-E/web_crawler.git/frontier"
)

type Product struct {
	Url, Image, Name, Price string
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func fetchPage(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", urlStr, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-Ok status code for %s: %d %s", urlStr, resp.StatusCode, resp.Status)
	}

	return resp, nil
}

func getTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}
	var buf strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			buf.WriteString(c.Data)
		} else if c.Type == html.ElementNode {
			buf.WriteString(getTextContent(c))
		}
	}
	return strings.TrimSpace(buf.String())
}

func processNode(n *html.Node, f *frontier.Frontier, baseURL *url.URL) {
	// process current node based on its type and attributes
	switch n.Type {
	case html.ElementNode:
		switch n.Data {
		case "h2":
			// check if FirstChild node of the h2 element is a text
			if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
				name := n.FirstChild.Data

				if name != "" {
					fmt.Println("Name:", strings.TrimSpace(name))
				}
			}
		case "span":
			// check for the span with class "amount"
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "amount") {
					// retrieve the text content of the "amount" span
					price := getTextContent(n)
					if price != "" {
						fmt.Println("Price:", price)
					}
				}
			}
		case "img":
			// check for the src attribute in the img tag
			for _, a := range n.Attr {
				if a.Key == "src" {
					// retrieve src value
					imageURL := a.Val
					fmt.Println("Image URL:", imageURL)

					parsedURL, err := baseURL.Parse(imageURL)
					if err == nil && parsedURL.IsAbs() {
						f.Add(parsedURL)
					}
				}
			}
		case "a", "link":
			for _, a := range n.Attr {
				if a.Key == "href" {
					linkURL := a.Val
					fmt.Printf("Found Link (href): %s\n", linkURL)
					parsedURL, err := baseURL.Parse(linkURL)
					if err == nil && parsedURL.IsAbs() {
						f.Add(parsedURL)
					}
				}
			}
		case "script", "iframe", "video", "audio":
			for _, a := range n.Attr {
				if a.Key == "src" {
					resourceURL := a.Val
					fmt.Printf("Found Resource (src): %s\n", resourceURL)
					parsedURL, err := baseURL.Parse(resourceURL)
					if err == nil && parsedURL.IsAbs() {
						f.Add(parsedURL)
					}
				}
			}
		case "form":
			for _, a := range n.Attr {
				if a.Key == "action" {
					actionURL := a.Val
					fmt.Printf("Found form Action: %s\n", actionURL)
					parsedURL, err := baseURL.Parse(actionURL)
					if err == nil && parsedURL.IsAbs() {
						f.Add(parsedURL)
					}
				}
			}
		}
	}

	// traverse child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(c, f, baseURL)
	}
}

func crawlerWorker(id int, f *frontier.Frontier) {
	// signal to the WaitGroup that this worker is starting
	f.AddWorker()
	// Ensure DoneWorker is called when the goroutine exits
	defer f.DoneWorker()

	for {
		currentURL := f.GetURL()
		if currentURL == nil {
			fmt.Printf("Wroker %d: Frontier closed or empty. Exiting.\n", id)
			return
		}

		fmt.Printf("Worker %d: Fetching %s\n", id, currentURL.String())
		resp, err := fetchPage(currentURL.String())
		if err != nil {
			fmt.Printf("Worker %d: Error fetching %s: %v\n", id, currentURL.String(), err)
			time.Sleep(1 * time.Second) // Small delay on error to avoid hammering
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		if err != nil {
			fmt.Printf("Worker %d: Error reading response body for %s: %v\n", id, currentURL.String(), err)
		}

		reader := bytes.NewReader(bodyBytes)
		parsedResp, err := html.Parse(reader)
		if err != nil {
			fmt.Printf("Worker %d: Error parsing HTML for %s: %v\n", id, currentURL.String(), err)
			continue
		}

		// Process the parsed HTML, extract data, and add new links to the frontier.
		// Use currentURL as the base for resolving relative links found on this page.
		processNode(parsedResp, f, currentURL)

		// Simulate some processing time to avoid overwhelming the server and
		// to make the concurrent nature visible.
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	seedURLStr := "https://www.scrapingcourse.com/ecommerce/"
	seedURL, err := url.Parse(seedURLStr)
	if err != nil {
		fmt.Println("Error parsing seed URL:", err)
		return
	}

	// Define initial URLs for the frontier and any exclusion patterns
	initialUrls := []url.URL{*seedURL}
	excludePatterns := []string{
		// Add any URLs or patterns you want to exclude from crawling.
		// For example: "example.com/admin", "google.com", ".pdf"
	}

	// Initialize the Frontier
	f := frontier.NewFrontier(initialUrls, excludePatterns)

	// Number of concurrent crawler workers
	numWorkers := 5
	fmt.Printf("Starting %d crawler workers...\n", numWorkers)
	for i := 1; i <= numWorkers; i++ {
		go crawlerWorker(i, f)
	}

	// --- Main Crawler Control Logic ---
	// Let the crawler run for a fixed duration for this example.
	// In a real application, you'd have more sophisticated termination logic:
	// - Max number of pages crawled
	// - No new URLs found for a certain period
	// - User initiated stop
	crawlDuration := 30 * time.Second // Run for 30 seconds
	fmt.Printf("Crawler will run for %s. Press Ctrl+C to stop earlier.\n", crawlDuration)
	time.Sleep(crawlDuration)

	fmt.Println("\nMain: Signaling frontier to terminate...")
	f.Terminate() // Gracefully shut down the frontier and wait for all workers to finish

	fmt.Println("Main: Crawler finished all active tasks.")

	// --- Optional: Render HTML to file for initial seed (kept from your original code) ---
	// This part is outside the main crawling loop and is just for saving the *initial* page's HTML.
	fmt.Println("\n--- Saving initial page HTML to output.html ---")
	resp, err := fetchPage(seedURLStr)
	if err != nil {
		fmt.Println("Error fetching initial URL for rendering:", err)
		return
	}
	bodyBytes, err := io.ReadAll(resp.Body) // Re-read bodyBytes for rendering
	resp.Body.Close()
	if err != nil {
		fmt.Println("Error reading initial response for rendering:", err)
		return
	}
	reader := bytes.NewReader(bodyBytes)
	parsedResp, err := html.Parse(reader)
	if err != nil {
		fmt.Println("Error parsing initial response for rendering:", err)
		return
	}
	var renderedHtml bytes.Buffer
	err = html.Render(&renderedHtml, parsedResp)
	if err != nil {
		fmt.Println("Error rendering HTML:", err)
		return
	}
	err = os.WriteFile("output.html", renderedHtml.Bytes(), 0644)
	if err != nil {
		fmt.Println("Error writing HTML to file:", err)
	} else {
		fmt.Println("HTML saved to output.html")
	}
}
