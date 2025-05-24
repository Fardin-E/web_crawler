package frontier

import (
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type Frontier struct {
	urls        chan *url.URL
	history     map[string]time.Time
	exclude     []string
	mu          sync.RWMutex
	wg          sync.WaitGroup
	closing     chan struct{}
	terminating bool
}

func NewFrontier(initialUrls []url.URL, exclude []string) *Frontier {
	f := &Frontier{
		urls:    make(chan *url.URL, len(initialUrls)*2),
		history: make(map[string]time.Time),
		exclude: exclude,
		mu:      sync.RWMutex{},
		wg:      sync.WaitGroup{},
		closing: make(chan struct{}),
	}

	for _, u := range initialUrls {
		f.Add(&u)
	}
	return f
}

func (f *Frontier) Add(u *url.URL) bool {
	url := u.String()

	f.mu.RLock()
	if f.terminating {
		f.mu.RUnlock()
		return false
	}
	if f.Seen(url) {
		f.mu.RUnlock()
		log.WithFields(log.Fields{"url": url}).Info("Already seen")
		return false
	}
	for _, pattern := range f.exclude {
		if strings.Contains(url, pattern) {
			f.mu.RUnlock()
			log.WithFields(log.Fields{"url": url}).Info("Excluded")
			return false
		}
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if f.terminating || f.Seen(url) { // Check again for termination or if it was added by another goroutine
		return false
	}

	f.history[url] = time.Now()

	select {
	case f.urls <- u:
		return true
	case <-f.closing: // Check for closing signal here
		log.WithFields(log.Fields{"url": url}).Warn("Frontier is closing, discarding URL.")
		delete(f.history, url) // Remove from history if not added
		return false
	case <-time.After(50 * time.Millisecond):
		log.WithFields(log.Fields{"url": url}).Warn("Frontier channel full, discarding URL.")
		delete(f.history, url)
		return false
	}
}

func (f *Frontier) GetURL() *url.URL {
	select {
	case u, ok := <-f.urls:
		if !ok {
			return nil
		}
		return u
	case <-f.closing:
		return nil
	}
}

func (f *Frontier) Terminate() {
	f.mu.Lock()
	if f.terminating {
		f.mu.Unlock()
		return
	}
	f.terminating = true
	close(f.closing)
	f.mu.Unlock()

	time.Sleep(100 * time.Millisecond)

	close(f.urls)

	log.Info("Frontier: Signaled termination. waiting for workers to finish")
	f.wg.Wait()
	log.Info("Frontier: All workes finished. Termminated.")
}

func (f *Frontier) Seen(urlStr string) bool {
	if lastFetch, ok := f.history[urlStr]; ok {
		return time.Since(lastFetch) < 2*time.Hour
	}
	return false
}

func (f *Frontier) AddWorker() {
	f.wg.Add(1)
}

func (f *Frontier) DoneWorker() {
	f.wg.Done()
}
