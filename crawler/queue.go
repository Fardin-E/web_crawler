package crawler

import (
	"math/rand"
	"net/url"

	"github.com/Fardin-E/web_crawler.git/frontier"

	log "github.com/sirupsen/logrus"
)

func distributeUrls(frontier *frontier.Frontier, distributedInputs []chan *url.URL) {
	HostToWorker := make(map[string]int)
	for url := range frontier.Get() {
		index := rand.Intn(len(distributedInputs))
		if prevIndex, ok := HostToWorker[url.Host]; ok {
			index = prevIndex
		} else {
			HostToWorker[url.Host] = index
		}
		distributedInputs[index] <- url
	}
}

func mergeResults(workerResults []chan CrawlResult, out chan CrawlResult) {
	collect := func(in chan CrawlResult) {
		for result := range in {
			out <- result
		}
		log.Println("Worker finished")
	}

	for i, result := range workerResults {
		log.Printf("Start collecting results from worker %d", i)
		go collect(result)
	}
}
