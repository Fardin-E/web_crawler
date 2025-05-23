package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	// "golang.org/x/net/html" // Not needed if goquery is being used
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

	req.Header.Set("User-Agent", "YourGoCrawler/1.0 (contact@example.com)") // Good practice

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch URL %s: %w", urlStr, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close() // Crucial: close body if returning nil error
		return nil, fmt.Errorf("Received non-Ok status code for %s: %d %s", urlStr, resp.StatusCode, resp.Status)
	}

	return resp, nil
}

// getTextContent (kept for reference)
// func getTextContent(n *html.Node) string {
// 	if n == nil {
// 		return ""
// 	}
// 	var buf strings.Builder
// 	for c := n.FirstChild; c != nil; c = c.NextSibling {
// 		if c.Type == html.TextNode {
// 			buf.WriteString(c.Data)
// 		} else if c.Type == html.ElementNode {
// 			buf.WriteString(getTextContent(c))
// 		}
// 	}
// 	return strings.TrimSpace(buf.String())
// }

func main() {
	targetUrl := "https://www.scrapingcourse.com/ecommerce/"
	resp, err := fetchPage(targetUrl)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	// This ensures the data is available for both saving to file and parsing with goquery.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	// 1. Save the raw HTML to file from bodyBytes
	err = os.WriteFile("output_raw.html", bodyBytes, 0644)
	if err != nil {
		log.Printf("Error writing raw HTML to file: %v", err)
	} else {
		log.Println("Raw HTML saved to output_raw.html")
	}

	// 2. Create a NEW reader from bodyBytes for goquery.
	// This reader starts at the beginning of the byte slice.
	readerForGoquery := bytes.NewReader(bodyBytes)

	doc, err := goquery.NewDocumentFromReader(readerForGoquery)

	if err != nil {
		log.Fatalf("Error parsing HTML with goquery: %v", err)
	}

	// --- Goquery-based Data Extraction ---
	var allProducts []Product
	log.Println("--- Starting Goquery Data Extraction ---")

	doc.Find("ul.products li.product").Each(func(i int, s *goquery.Selection) {
		product := Product{}
		product.Url, _ = s.Find("a.woocommerce-LoopProduct-link").Attr("href")
		product.Name = strings.TrimSpace(s.Find("h2.woocommerce-loop-product__title").Text())
		product.Price = strings.TrimSpace(s.Find("span.product-price.woocommerce-Price-amount.amount").Text())
		product.Image, _ = s.Find("img.wp-post-image").Attr("src")
		allProducts = append(allProducts, product)
	})

	log.Println("--- Goquery Data Extraction Finished ---")

	fmt.Println("\n--- Scraped Products ---")
	if len(allProducts) == 0 {
		fmt.Println("No products found. Please verify your CSS selectors against the website's HTML structure.")
	}
	for i, p := range allProducts {
		fmt.Printf("Product %d:\n", i+1)
		fmt.Printf("  Name:  %s\n", p.Name)
		fmt.Printf("  Price: %s\n", p.Price)
		fmt.Printf("  Image: %s\n", p.Image)
		fmt.Printf("  URL:   %s\n", p.Url)
		fmt.Println("--------------------")
	}
}
