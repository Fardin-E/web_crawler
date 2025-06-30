package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Fardin-E/web_crawler.git/parser"
	log "github.com/sirupsen/logrus"
)

type CrawlResult struct {
	Url          *url.URL
	StatusCode   int
	ContentType  string
	ResponseTime time.Duration
	Body         []byte
	Info         *parser.Info
	IsError      bool
}

type Worker struct {
	input      chan *url.URL
	deadLetter chan *url.URL
	result     chan CrawlResult
	done       chan struct{}
	id         int
	logger     *log.Entry

	// Only contains the host part of the URL
	history map[string]time.Time
}

func NewWorker(input chan *url.URL, result chan CrawlResult, done chan struct{}, id int, deadLetter chan *url.URL) *Worker {
	history := make(map[string]time.Time)
	logger := log.WithField("worker", id)
	return &Worker{
		input:      input,
		result:     result,
		done:       done,
		id:         id,
		history:    history,
		deadLetter: deadLetter,
		logger:     logger,
	}
}
func (w *Worker) Start() {
	w.logger.Debugf("Worker %d started", w.id)
	for {
		select {
		case url := <-w.input:
			content, err := w.fetch(url)
			if err != nil {
				log.Errorf("Worker %d error fetching content: %s", w.id, err)
				w.deadLetter <- url
				continue
			}
			w.result <- content
		case <-w.done:
			return
		}
	}
}

func (w *Worker) CheckPoliteness(url *url.URL) bool {
	if lastFetch, ok := w.history[url.Host]; ok {
		return time.Since(lastFetch) > 2*time.Second
	}
	return true
}

func (w *Worker) fetch(url *url.URL) (CrawlResult, error) {
	w.logger.Debugf("Worker %d fetching %s", w.id, url)
	defer func() {
		w.history[url.Host] = time.Now()
	}()
	for !w.CheckPoliteness(url) {
		time.Sleep(2 * time.Second)
	}

	// Measure response time
	start := time.Now()
	res, err := http.Get(url.String())
	if err != nil {
		return CrawlResult{Url: url, IsError: true}, err
	}
	defer res.Body.Close()
	responseTime := time.Since(start)

	// Read body first, then check status code, so body is available for error logging/parsing
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return CrawlResult{
			Url:          url,
			StatusCode:   res.StatusCode,
			ContentType:  res.Header.Get("Content-Type"),
			ResponseTime: responseTime,
			IsError:      true,
		}, fmt.Errorf("failed to read response body for %s: %w", url.String(), err)
	}

	// Content Type Inference
	inferredContentType := res.Header.Get("Content-Type") // Get is safer, returns "" if not found
	if inferredContentType == "" {
		inferredContentType = http.DetectContentType(body)
	}

	// Return partial CrawlResult; Info will be populated later
	return CrawlResult{
		Url:          url,
		StatusCode:   res.StatusCode,
		ContentType:  inferredContentType,
		Body:         body,
		ResponseTime: responseTime,
		IsError:      false, // If we reached here, it's not an HTTP error
		Info:         nil,   // Info will be populated by a separate parsing step
	}, nil
}
