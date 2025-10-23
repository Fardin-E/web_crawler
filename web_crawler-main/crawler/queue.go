package crawler

import (
	"math/rand"
	"net/url"
	"sync"

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

	// Close all worker input channels when frontier is exhausted
	log.Debug("Frontier exhausted, closing worker input channels")
	for _, ch := range distributedInputs {
		close(ch)
	}
}

func mergeResults(workerResults []chan CrawlResult, out chan CrawlResult) {
	var wg sync.WaitGroup

	collect := func(in chan CrawlResult) {
		defer wg.Done()
		for result := range in {
			out <- result
		}
		log.Debug("Worker finished sending results")
	}

	for i, result := range workerResults {
		log.Printf("Start collecting results from worker %d", i)
		wg.Add(1)
		go collect(result)
	}

	// Wait for all workers to finish, then close output channel
	go func() {
		wg.Wait()
		log.Debug("All workers finished, closing merged results channel")
		close(out)
	}()
}
