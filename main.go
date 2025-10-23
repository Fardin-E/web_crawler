package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Fardin-E/web_crawler.git/crawler"
	"github.com/Fardin-E/web_crawler.git/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose bool
	config  string

	// Crawl command flags
	urls            []string
	depth           int
	workers         int
	outputDir       string
	excludePatterns []string
	revisitDelay    time.Duration
	maxRedirects    int

	// Serve command flags
	port int
	host string
)

// MAIN ENTRY POINT

func main() {
	// Root command
	var rootCmd = &cobra.Command{
		Use:   "crawler",
		Short: "A concurrent web crawler built in Go",
		Long: `A concurrent web crawler that can extract and process content from websites.
		Supports multiple storage backends, custom processors, and rate limiting`,
	}

	// Global flags availabl to all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "Config file path (default is ./config.yaml)")

	// Add subcommands
	rootCmd.AddCommand(crawlCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(versionCmd())

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}

// CRAWL COMMAND

func crawlCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crawl",
		Short: "Start crawling specific URLs",
		Long: `Start a web crawl from one or more seed URLs.
Examples:
  # Crawl a single URL
  crawler crawl --url https://example.com

  # Crawl multiple URLs with custom settings
  crawler crawl --url https://example.com --url https://another.com --workers 20 --depth 5

  # Crawl with exclude patterns
  crawler crawl --url https://example.com --exclude /login --exclude /admin
`,
		RunE: runCrawl,
	}

	// Add flags specific to crawl command
	cmd.Flags().StringSliceVarP(&urls, "url", "u", []string{}, "URL(s) to crawl (required, can be specified multiple times)")
	cmd.Flags().IntVarP(&depth, "depth", "d", 3, "Maximum crawl depth")
	cmd.Flags().IntVarP(&workers, "workers", "w", 10, "Number of concurrent workers")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "./data", "Output directory for crawled data")
	cmd.Flags().StringSliceVarP(&excludePatterns, "exclude", "e", []string{}, "URL patterns to exclude (can be specified multiple times)")
	cmd.Flags().DurationVar(&revisitDelay, "revisit-delay", 2*time.Hour, "Delay before revisiting a URL")
	cmd.Flags().IntVar(&maxRedirects, "max-redirects", 5, "Maximum number of redirects to follow")

	// Mark required flags
	cmd.MarkFlagRequired("url")

	return cmd
}

func runCrawl(cmd *cobra.Command, args []string) error {
	// Setup logging
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	log.Info("Starting web crawler...")
	log.Infof("URLs: %v", urls)
	log.Infof("Workers: %d", workers)
	log.Infof("Output directory: %s", outputDir)

	// Parse URLs
	initialUrls := []url.URL{}
	for _, urlStr := range urls {
		parsedUrl, err := url.Parse(urlStr)
		if err != nil {
			return fmt.Errorf("invalid URL '%s' : %w", urlStr, err)
		}
		initialUrls = append(initialUrls, *parsedUrl)
	}

	// Create storage
	contentStorage, err := storage.NewFileStorage(outputDir)
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Create crawler config
	crawlerConfig := &crawler.Config{
		MaxRedirects:    maxRedirects,
		RevisitDelay:    revisitDelay,
		WorkerCount:     workers,
		ExcludePatterns: excludePatterns,
	}

	// Create crawler
	c := crawler.NewCrawler(initialUrls, contentStorage, crawlerConfig)

	// Add custom processors
	c.AddProcessor(&LoggerProcessor{})

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start crawler in goroutine
	done := make(chan struct{})
	go func() {
		c.Start()
		close(done)
	}()

	// Wait for completion or interrupt
	select {
	case <-sigChan:
		log.Info("Received interrupt signal, shutting down gracefully.....")
		c.Terminate()
		<-done
		log.Info("Crawler stopped")
	case <-done:
		log.Info("Crawl completed")
	}

	return nil
}

// SERVE COMMAND

func serveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the crawler API server",
		Long: `Start an HTTP API server to manage crawl jobs remotely.

Examples:
  # Start server on default port
  crawler serve
  
  # Start on custom port
  crawler serve --port 3000
  
  # Start on specific host and port
  crawler serve --host 0.0.0.0 --port 8080 
`,
		RunE: runServe,
	}

	cmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to listen on")
	cmd.Flags().StringVar(&host, "host", "localHost", "Host to bind to")

	return cmd
}

func runServe(cmd *cobra.Command, args []string) error {
	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	address := fmt.Sprintf("%s:%d", host, port)
	log.Infof("Starting API server on %s", address)

	// Setup routes
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/api/v1/crawl", crawlJobHandler)

	log.Infof("API server running at http://%s", address)
	log.Info("Endpoints:")
	log.Info("  GET   /health         - Health check")
	log.Info("  POST  /api/v1/crawl   - Start crawl job")

	return http.ListenAndServe(address, nil)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","version":"1.0.0"}`))
}

func crawlJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// TODO: Implement crawl job creation
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"job_id":"123","status":"queued"}`))
}

// VERSION COMMAND

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Web crawler v1.0.0")
			fmt.Println("Built with Go")
		},
	}
}

// CUSTOM PROCESS

type LoggerProcessor struct{}

func (l *LoggerProcessor) Process(result *crawler.CrawlResult) error {
	log.WithFields(log.Fields{
		"url":          result.Url.String(),
		"content_type": result.ContentType,
		"size":         len(result.Body),
	}).Info("Processed page")
	return nil
}
