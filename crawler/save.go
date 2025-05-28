package crawler

import (
	"encoding/json"
	"net/url"
	"path"

	"github.com/Fardin-E/web_crawler.git/storage"
)

type SaveToFile struct {
	storageBackend storage.Storage
}

func (s *SaveToFile) Process(result CrawlResult) error {
	savePath := getSavePath(result.Url)

	switch result.ContentType {
	default:
		savePath := savePath + ".json"

		jsonData, err := json.MarshalIndent(result, "", " ")
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
