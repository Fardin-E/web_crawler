package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/Fardin-E/web_crawler.git/fetcher"
	"github.com/Fardin-E/web_crawler.git/parser"
)

func main() {
	url := "https://www.scrapingcourse.com/ecommerce/"

	fetcher := &fetcher.Fetcher{}

	fmt.Printf("Fetching HTML from: %s", url)
	resp, err := fetcher.FetchPage(url)
	if err != nil {
		fmt.Printf("Error fetching url %s\n", url)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body %v\n", err)
		return
	}

	// --- START BOM STRIPPING AND RAW BYTE INSPECTION ---
	fmt.Printf("DEBUG: Fetched body length (raw): %d bytes\n", len(bodyBytes))
	if len(bodyBytes) > 10 {
		fmt.Printf("DEBUG: First 10 raw bytes (hex): %s\n", hex.EncodeToString(bodyBytes[:10]))
	} else {
		fmt.Printf("DEBUG: Raw bytes (hex): %s\n", hex.EncodeToString(bodyBytes))
	}

	// Check for UTF-8 BOM (EF BB BF) and strip it if present
	if len(bodyBytes) >= 3 && bytes.Equal(bodyBytes[0:3], []byte{0xEF, 0xBB, 0xBF}) {
		fmt.Println("DEBUG: UTF-8 BOM detected and stripped.")
		bodyBytes = bodyBytes[3:]
	}
	// --- END BOM STRIPPING AND RAW BYTE INSPECTION ---

	htmlContent := string(bodyBytes) // Convert bytes to string AFTER BOM stripping

	parser := &parser.HtmlParser{}
	parsedTokens, err := parser.Parse(htmlContent)
	if err != nil {
		fmt.Printf("Error parsing HTML: %v\n", err)
		return
	}

	fmt.Println("\nExtracted Tokens:")
	if len(parsedTokens) == 0 {
		fmt.Println("  No tokens extracted.")
	}
	for _, t := range parsedTokens {
		fmt.Printf("  Name: %-8s Value: %s\n", t.Name, t.Value)
	}
}
