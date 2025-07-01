package crawler

import (
	"context"
	"fmt"
	"time"

	"github.com/Fardin-E/web_crawler.git/backend/storage"
)

type SaveToDB struct {
	dbstorage storage.ConnDB
}

func (s *SaveToDB) Process(result *CrawlResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	pageURL := result.Url.String()
	statusCode := result.StatusCode
	contentType := result.ContentType
	contentLength := len(result.Body)
	responseTimeMs := int(result.ResponseTime.Milliseconds())
	isError := result.IsError
	rawHtml := string(result.Body) // Extract raw HTML directly

	var title string
	var metaDescription string
	var outLinks []string
	var paragraphs []string // Declare paragraphs here

	if result.Info != nil {
		title = result.Info.Title
		metaDescription = result.Info.Description
		// convert []parser.Token to []string for out_links
		for _, linkToken := range result.Info.Links {
			outLinks = append(outLinks, linkToken.Value)
		}
		paragraphs = result.Info.Paragraphs // Assign parsed paragraphs here
	}

	err := s.dbstorage.InsertPage(
		ctx,
		pageURL,
		statusCode,
		contentType,
		// htmlContentJSON, // REMOVED: No longer passing the JSON blob
		title,           // Direct column
		metaDescription, // Direct column
		contentLength,
		responseTimeMs,
		outLinks, // Direct column TEXT[]
		isError,
		rawHtml,    // ADDED: Pass raw HTML
		paragraphs, // ADDED: Pass paragraphs
	)
	if err != nil {
		return fmt.Errorf("failed to insert page into DB: %w", err)
	}
	return nil
}
