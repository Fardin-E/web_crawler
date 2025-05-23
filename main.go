package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
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
		return nil, fmt.Errorf("Failed to fetch URL %s: %w", urlStr, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received non-Ok status code for %s: %d %s", urlStr, resp.StatusCode, resp.Status)
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

func processNode(n *html.Node) {
	switch n.Data {
	case "h2":
		// check if FirstChild node of the h2 element is a text
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			// if yes, retrieve FirstChild's data (name)
			name := n.FirstChild.Data
			// print name
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
				ImageUrl := a.Val
				// print image URL
				fmt.Println("Image URL:", ImageUrl)
			}
		}
	}

	// Traverse child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(c)
	}
}

func main() {
	targetUrl := "https://www.scrapingcourse.com/ecommerce/"
	resp, err := fetchPage(targetUrl)
	if err != nil {
		fmt.Println("Error fetching URL:", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error parsing response", err)
		return
	}

	reader := bytes.NewReader(bodyBytes)
	parsedResp, err := html.Parse(reader)
	if err != nil {
		fmt.Println("Error parsing response:", err)
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

	fmt.Println("--- Starting Data Extraction ---")
	processNode(parsedResp)
	fmt.Println("--- Data Extraction Finished ---")
}
