package main

import (
	"net/url"
	"time"

	"github.com/Fardin-E/web_crawler.git/crawler"
	"github.com/Fardin-E/web_crawler.git/storage"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	initialUrls := []url.URL{}

	myUrl, _ := url.Parse("https://www.bbc.com/news/articles/cql2990k4dno")
	initialUrls = append(initialUrls, *myUrl)

	contentStorage, err := storage.NewFileStorage("./data")
	if err != nil {
		panic(err)
	}

	// contentParsers := []parser.Parser{}
	// contentParsers = append(contentParsers, &parser.HtmlParser{})

	skipPatterns := []string{"/login", "/search", "/cart", "/checkout"}

	crawler := crawler.NewCrawler(initialUrls, contentStorage, &crawler.Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour * 2,
		WorkerCount:     10,
		ExcludePatterns: skipPatterns,
	})

	// adding custom parser to the crawler

	// Adding custom processor to the crawler
	crawler.AddProcessor(&LoggerProcessor{})

	crawler.Start()
}

// Example of custom processor
type LoggerProcessor struct {
}

func (l *LoggerProcessor) Process(result crawler.CrawlResult) error {
	log.Print("Processing result")
	return nil
}
