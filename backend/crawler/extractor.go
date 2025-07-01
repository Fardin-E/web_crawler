package crawler

import (
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type LinkExtractor struct {
	NewUrls chan *url.URL
}

func (e *LinkExtractor) Process(result *CrawlResult) error {
	if result.Info == nil {
		return fmt.Errorf("no parsed Info available for URL: %s", result.Url)
	}

	foundUrls := make([]*url.URL, 0)
	for _, parsedUrl := range result.Info.Links {
		newUrl, err := url.Parse(parsedUrl.Value)
		if err != nil {
			log.Debugf("Error parsing url: %s", err)
			continue
		}
		params := newUrl.Query()
		for param := range params {
			newUrl = stripQueryParam(newUrl, param)
		}
		if newUrl.Scheme == "http" || newUrl.Scheme == "https" {
			foundUrls = append(foundUrls, newUrl)
		}
	}
	log.Infof("Extracted %d urls", len(foundUrls))
	for _, foundUrl := range foundUrls {
		e.NewUrls <- foundUrl
	}
	return nil
}

func stripQueryParam(inputURL *url.URL, stripKey string) *url.URL {
	query := inputURL.Query()
	query.Del(stripKey)
	inputURL.RawQuery = query.Encode()
	return inputURL
}
