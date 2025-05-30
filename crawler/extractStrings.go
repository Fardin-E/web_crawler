package crawler

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type HTMLProcessor struct{}

func NewHTMLProcessor() *HTMLProcessor {
	return &HTMLProcessor{}
}

// Helper function to detect JavaScript/JSON content
func isJavaScriptOrJSON(text string) bool {
	// Check for common JS/JSON patterns
	jsPatterns := []string{
		"function", "var ", "const ", "let ", "window.", "document.",
		"{\"", "\":\"", "\\/", "\\u", "wp.", "jQuery", "lodash",
	}

	for _, pattern := range jsPatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// Filter individual words that look like JS artifacts
func containsJSArtifacts(word string) bool {
	// Skip words that are clearly JS/JSON artifacts
	if strings.Contains(word, "{") || strings.Contains(word, "}") ||
		strings.Contains(word, "\":") || strings.Contains(word, "\\/") ||
		strings.Contains(word, "\\u") || len(word) > 50 {
		return true
	}
	return false
}

func (h *HTMLProcessor) ProcessNode(n *html.Node) ([]byte, error) {
	bodyNode := h.findBodyNode(n)
	if bodyNode == nil {
		return []byte{}, fmt.Errorf("body element not found")
	}

	words := h.extractStrings(bodyNode)
	result := strings.Join(words, " ")
	return []byte(result), nil
}

// Helper function to find the body element
func (h *HTMLProcessor) findBodyNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if result := h.findBodyNode(child); result != nil {
			return result
		}
	}
	return nil
}

func (h *HTMLProcessor) extractStrings(n *html.Node) []string {
	var words []string

	var traverse func(*html.Node) bool
	traverse = func(node *html.Node) bool {
		if len(words) >= 500 {
			return false
		}

		// Skip script, style, and other non-content elements
		if node.Type == html.ElementNode {
			switch strings.ToLower(node.Data) {
			case "script", "style", "noscript", "meta", "link", "head":
				return true // Skip this entire subtree
			}
		}

		if node.Type == html.TextNode {
			text := strings.TrimSpace(node.Data)
			if text != "" && !isJavaScriptOrJSON(text) {
				nodeWords := strings.Fields(text)
				for _, word := range nodeWords {
					// Filter out obvious JavaScript/JSON artifacts
					if !containsJSArtifacts(word) {
						words = append(words, word)
						if len(words) >= 500 {
							return false
						}
					}
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			if !traverse(child) {
				return false
			}
		}
		return true
	}

	traverse(n)
	return words
}
