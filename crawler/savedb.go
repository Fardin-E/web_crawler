package crawler

import (
	"context"
	"encoding/json"
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

	jsonData, err := json.Marshal(result.Info)
	if err != nil {
		return fmt.Errorf("failed to marshal Info to JSON: %w", err)
	}

	err = s.dbstorage.InsertPage(ctx, result.Url.String(), result.Info.StatusCode, result.ContentType, jsonData)
	if err != nil {
		return fmt.Errorf("failed to insert page: %w", err)
	}
	return nil
}
