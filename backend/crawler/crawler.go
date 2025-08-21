package crawler

import (
	"net/url"
	"time"

	"github.com/Fardin-E/web_crawler.git/backend/frontier"
	"github.com/Fardin-E/web_crawler.git/backend/parser"
	"github.com/Fardin-E/web_crawler.git/backend/storage"

	log "github.com/sirupsen/logrus"
)

type Crawler struct {
	config         *Config
	frontier       *frontier.Frontier
	storage        storage.Storage
	contentParsers []parser.Parser
	deadLetter     chan *url.URL
	processors     []Processor
}

func NewCrawler(initialUrls []url.URL,
	contentStorage storage.Storage,
	config *Config) *Crawler {
	deadLetter := make(chan *url.URL)
	contentParser := []parser.Parser{&parser.HtmlParser{}}
	return &Crawler{
		frontier:       frontier.NewFrontier(initialUrls, config.ExcludePatterns),
		storage:        contentStorage,
		contentParsers: contentParser,
		deadLetter:     deadLetter,
		config:         config,
	}
}

func (c *Crawler) Start() {
	conn := &storage.ConnDB{}
	err := conn.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	distributedInputs := make([]chan *url.URL, c.config.WorkerCount)
	workersResults := make([]chan CrawlResult, c.config.WorkerCount)
	done := make(chan struct{})

	for i := range c.config.WorkerCount {
		distributedInputs[i] = make(chan *url.URL)
		workersResults[i] = make(chan CrawlResult)
	}
	go distributeUrls(c.frontier, distributedInputs)
	for i := range c.config.WorkerCount {
		worker := NewWorker(distributedInputs[i], workersResults[i], done, i, c.deadLetter)
		go worker.Start()
	}

	mergedResults := make(chan CrawlResult)
	go mergeResults(workersResults, mergedResults)
	newUrls := make(chan *url.URL)
	c.AddProcessor(&LinkExtractor{NewUrls: newUrls})
	c.AddProcessor(&SaveToFile{storageBackend: c.storage})
	c.AddProcessor(&SaveToDB{dbstorage: *conn})
	go func() {
		for newUrl := range newUrls {
			_ = c.frontier.Add(newUrl)
		}
	}()

	go func() {
		for deadUrl := range c.deadLetter {
			log.Debugf("Dismissed %s", deadUrl)
		}
	}()

	for result := range mergedResults {
		copyResult := result

		// Parse once BEFORE passing to processors
		for _, p := range c.contentParsers {
			if p.IsSupportedExtension(copyResult.ContentType) {
				parsedInfo, err := p.Parse(string(copyResult.Body))
				if err != nil {
					log.Warnf("Failed to parse: %v", err)
				} else {
					// 3. Assign the new parsed Info struct
					copyResult.Info = &parsedInfo
				}
				break // Use only the first matching parser
			}
		}

		for _, processor := range c.processors {
			go func(processor Processor, r *CrawlResult) {
				if err := processor.Process(r); err != nil {
					log.Error(err)
				}
			}(processor, &copyResult)
		}
	}
	log.Println("Crawler exited")
}

func (c *Crawler) Terminate() {
	c.frontier.Terminate()
}
func (c *Crawler) AddContentParser(contentParser parser.Parser) {
	c.contentParsers = append(c.contentParsers, contentParser)
}

func (c *Crawler) AddExcludePattern(pattern string) {
	c.config.ExcludePatterns = append(c.config.ExcludePatterns, pattern)
}

func (c *Crawler) AddProcessor(processor Processor) {
	c.processors = append(c.processors, processor)
}

// StartCrawling sets up and runs the crawler with a given starting URL.
func StartCrawling(startURL string) {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	myUrl, err := url.Parse(startURL)
	if err != nil {
		log.Printf("Error parsing URL: %v", err)
		return
	}

	initialUrls := []url.URL{*myUrl}

	contentStorage, err := storage.NewFileStorage("../data")
	if err != nil {
		log.Printf("Error creating storage: %v", err)
		return
	}

	skipPatterns := []string{"/login*", "/search*", "/cart*", "/checkout*", "/account*"}

	// Note: You might want to make these configs more dynamic
	crawler := NewCrawler(initialUrls, contentStorage, &Config{
		MaxRedirects:    5,
		RevisitDelay:    time.Hour * 2,
		WorkerCount:     10,
		ExcludePatterns: skipPatterns,
	})

	go crawler.Start()
}
