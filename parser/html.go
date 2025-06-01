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

func (p *HtmlParser) Parse(content string) (Info, error) {
	htmlParser := html.NewTokenizer(strings.NewReader(content))
	info := Info{
		Paragraphs: []string{},
		Links:      []Token{},
	}

	var collectingText bool
	var textBuffer strings.Builder

	for {
		tokenType := htmlParser.Next()
		if tokenType == html.ErrorToken {
			break
		}

		token := htmlParser.Token()

		switch tokenType {
		case html.StartTagToken, html.SelfClosingTagToken:
			tagName := strings.ToLower(token.Data)

			switch tagName {
			case "a":
				for _, attr := range token.Attr {
					if strings.ToLower(attr.Key) == "href" {
						info.Links = append(info.Links, Token{Name: "link", Value: attr.Val})
					}
				}

			case "title":
				collectingText = true
				textBuffer.Reset()

			case "meta":
				name, content, property := "", "", ""

				for _, attr := range token.Attr {
					switch strings.ToLower(attr.Key) {
					case "name":
						name = strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					case "property":
						property = strings.ToLower(attr.Val)
					}
				}

				// Extract meta description
				if (name == "description" || property == "og:description") && info.Description == "" {
					info.Description = content
				}

			case "h1", "h2", "h3", "h4", "h5", "h6", "p", "article", "main":
				collectingText = true
				textBuffer.Reset()
			}

		case html.TextToken:
			if collectingText {
				text := strings.TrimSpace(token.Data)
				if text != "" {
					if textBuffer.Len() > 0 {
						textBuffer.WriteString(" ")
					}
					textBuffer.WriteString(text)
				}
			}

		case html.EndTagToken:
			if !collectingText {
				continue
			}

			tagName := strings.ToLower(token.Data)
			text := strings.TrimSpace(textBuffer.String())

			switch tagName {
			case "title":
				if info.Title == "" && text != "" {
					info.Title = text
				}

			case "h1":
				// Use H1 as title fallback
				if info.Title == "" && text != "" {
					info.Title = text
				}
				// Also add to paragraphs
				if text != "" {
					info.Paragraphs = append(info.Paragraphs, text)
				}

			case "h2", "h3", "h4", "h5", "h6", "p":
				if text != "" {
					info.Paragraphs = append(info.Paragraphs, text)
				}

			case "article", "main":
				// Only add substantial content from article/main tags
				if text != "" && len(text) > 50 {
					info.Paragraphs = append(info.Paragraphs, text)
				}
			}

			collectingText = false
			textBuffer.Reset()
		}
	}

	// Limit total content by removing excess paragraphs
	totalLength := 0
	maxLength := 2000

	for i, paragraph := range info.Paragraphs {
		totalLength += len(paragraph)
		if totalLength > maxLength {
			// Truncate at this paragraph and add ellipsis
			if i > 0 {
				info.Paragraphs = info.Paragraphs[:i]
			} else {
				// If first paragraph is too long, truncate it
				info.Paragraphs = []string{paragraph[:maxLength-3] + "..."}
			}
			break
		}
	}

	return info, nil
}
