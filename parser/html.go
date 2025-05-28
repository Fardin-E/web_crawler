package parser

import (
	"strings"

	"golang.org/x/net/html"
)

type HtmlParser struct {
}

func (p *HtmlParser) getSupportedExtensions() []string {
	return []string{".html", ".htm"}
}

func (p *HtmlParser) IsSupportedExtension(extension string) bool {
	for _, supportedExtension := range p.getSupportedExtensions() {
		if extension == supportedExtension {
			return true
		}
	}
	return true
}

func (p *HtmlParser) Parse(content string) ([]Token, error) {
	htmlParser := html.NewTokenizer(strings.NewReader(content))
	tokens := []Token{}
	for {
		tokenType := htmlParser.Next()
		if tokenType == html.ErrorToken {
			break
		}
		token := htmlParser.Token()
		// Handle both start tags and self-closing tags
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
			switch strings.ToLower(token.Data) {
			case "a":
				for _, attr := range token.Attr {
					if strings.ToLower(attr.Key) == "href" {
						tokens = append(tokens, Token{Name: "link", Value: attr.Val})
					}
				}
			case "img":
				for _, attr := range token.Attr {
					if strings.ToLower(attr.Key) == "src" {
						tokens = append(tokens, Token{Name: "image", Value: attr.Val})
					}
				}
			}
		}
	}
	return tokens, nil
}
