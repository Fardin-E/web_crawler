package crawler

import (
	"net/url"

	"github.com/Fardin-E/web_crawler.git/frontier"
	"github.com/Fardin-E/web_crawler.git/parser"
	"github.com/Fardin-E/web_crawler.git/storage"

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
		// Parse once BEFORE passing to processors
		for _, parser := range c.contentParsers {
			if parser.IsSupportedExtension(result.ContentType) {
				parsedInfo, err := parser.Parse(string(result.Body))
				if err != nil {
					log.Warnf("Failed to parse: %v", err)
				} else {
					result.Info = &parsedInfo
				}
				break // Use only the first matching parser
			}
		}

		for _, processor := range c.processors {
			go func(processor Processor, result *CrawlResult) {
				if err := processor.Process(result); err != nil {
					log.Error(err)
				}
			}(processor, &result)
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
