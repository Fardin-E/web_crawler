package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Fetcher struct {
}

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func (f *Fetcher) FetchPage(url *url.URL) (*http.Response, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL %s: %w", url.String(), err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-Ok status code for %s: %d %s", url.String(), resp.StatusCode, resp.Status)
	}

	return resp, nil
}
