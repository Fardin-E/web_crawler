package crawler

import (
	"encoding/json"
	"net/url"
	"path"
	"strings"

	"github.com/Fardin-E/web_crawler.git/storage"
	"golang.org/x/net/html"
)

type SaveToFile struct {
	storageBackend storage.Storage
}

func (s *SaveToFile) Process(result CrawlResult) error {
	savePath := getSavePath(result.Url)

	switch result.ContentType {
	default:
		savePath := savePath + ".json"

		// Create JSON-friendly version with string body
		bodyText := string(result.Body) // Default to original body as string

		if strings.Contains(result.ContentType, "text/html") {
			doc, err := html.Parse(strings.NewReader(string(result.Body)))
			if err == nil {
				processor := NewHTMLProcessor()
				processedBody, err := processor.ProcessNode(doc)
				if err == nil {
					bodyText = string(processedBody) // Convert processed bytes to string
				}
			}
		}

		// Create struct for JSON serialization
		jsonResult := struct {
			Url         *url.URL `json:"url"`
			ContentType string   `json:"contentType"`
			Body        string   `json:"body"`
		}{
			Url:         result.Url,
			ContentType: result.ContentType,
			Body:        bodyText,
		}

		jsonData, err := json.MarshalIndent(jsonResult, "", " ")
		if err != nil {
			return err
		}

		err = s.storageBackend.Set(savePath, string(jsonData))
		if err != nil {
			return err
		}
	}
	return nil
}

func getSavePath(url *url.URL) string {
	fileName := url.Path
	savePath := path.Join(url.Host, fileName)
	return savePath
}
