package crawler

import (
	"context"
	"fmt"
	"time"

	"github.com/Fardin-E/web_crawler.git/storage"
)

type SaveToDB struct {
	dbstorage storage.ConnDB
}

func (s *SaveToDB) Process(result *CrawlResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := s.dbstorage.InsertPage(ctx, result.Url.String(), result.ContentType, string(result.Body))
	if err != nil {
		return fmt.Errorf("failed to insert page: %w", err)
	}
	return nil
}
