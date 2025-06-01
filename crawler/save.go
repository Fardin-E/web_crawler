package crawler

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/Fardin-E/web_crawler.git/storage"
)

type SaveToFile struct {
	storageBackend storage.Storage
}

var imageExtensions = map[string]string{
	"image/jpeg":    ".jpg",
	"image/png":     ".png",
	"image/gif":     ".gif",
	"image/webp":    ".webp",
	"image/svg+xml": ".svg",
	// ... etc
}

func getImageExtension(contentType string) string {
	if ext, ok := imageExtensions[contentType]; ok {
		return ext
	}
	return ".bin"
}

func (s *SaveToFile) Process(result CrawlResult) error {
	savePath := getSavePath(result.Url)

	switch {
	case strings.HasPrefix(result.ContentType, "text/html"):
		savePath := savePath + ".html"
		return s.storageBackend.Set(savePath, string(result.Body))

	case strings.HasPrefix(result.ContentType, "image/"):
		ext := getImageExtension(result.ContentType)
		savePath := savePath + ext
		return s.storageBackend.Set(savePath, string(result.Body))

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
