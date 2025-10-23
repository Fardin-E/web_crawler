package crawler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/Fardin-E/web_crawler.git/storage"
)

type SaveToFile struct {
	storageBackend storage.Storage
}

func (s *SaveToFile) Process(result *CrawlResult) error {
	savePath := getSavePath(result.Url)

	switch {
	case strings.HasPrefix(result.ContentType, "text/html"):
		if result.Info != nil {
			jsonPath := savePath + ".json"
			data, err := json.MarshalIndent(result.Info, "", "  ")
			if err != nil {
				return err
			}
			return s.storageBackend.Set(jsonPath, string(data))
		}

		return nil

	default:
		// Handle other content types or return error
		return fmt.Errorf("unsupported content type: %s", result.ContentType)
	}
}

func getSavePath(url *url.URL) string {
	fileName := url.Path
	savePath := path.Join(url.Host, fileName)
	return savePath
}
