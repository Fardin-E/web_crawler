## Web crawler made in Golang
# 🕷️ Web Crawler

A high-performance, concurrent web crawler built with Go, featuring configurable worker pools, intelligent URL deduplication, and Docker containerization.

## ✨ Features

- ⚡ **Concurrent Crawling** - Multi-threaded architecture with configurable worker pools
- 🔄 **Smart URL Management** - Automatic deduplication and revisit control
- 🎯 **Pattern Exclusion** - Filter unwanted domains (ads, trackers, etc.)
- 🤝 **Politeness Delay** - Respects server resources with configurable delays
- 📦 **Storage System** - Persistent file-based content storage
- 🐳 **Docker Ready** - Fully containerized with multi-stage builds
- 🧪 **Well Tested** - Comprehensive test suite with 75% code coverage
- 🔄 **CI/CD Pipeline** - Automated testing and building with GitHub Actions

## 🚀 Quick Start

### Using Docker (Recommended)

```bash
# Pull and run
docker run --rm web-crawler:latest crawl --url https://example.com --workers 5

# With persistent storage
docker run --rm -v $(pwd)/data:/home/crawler/data \
  web-crawler:latest crawl --url https://golang.org
```

### Using Go

```bash
# Clone the repository
git clone https://github.com/Fardin-E/web_crawler.git
cd web_crawler

# Install dependencies
go mod download

# Run the crawler
go run . crawl --url https://example.com --workers 10 --verbose
```

### Using Docker Compose

```bash
# Start crawler
docker-compose up crawler

# Or run in background
docker-compose up -d crawler
```

## 📖 Usage

### Basic Crawling

```bash
# Crawl a single URL
./crawler crawl --url https://example.com

# Crawl with custom settings
./crawler crawl \
  --url https://example.com \
  --workers 20 \
  --output ./data \
  --verbose

# Crawl multiple URLs
./crawler crawl \
  --url https://site1.com \
  --url https://site2.com \
  --workers 10
```

### API Server Mode

```bash
# Start API server
./crawler serve --port 8080 --host 0.0.0.0

# Then access at http://localhost:8080
```

### Configuration Options

| Flag | Description | Default |
|------|-------------|---------|
| `--url` | URL(s) to crawl | Required |
| `--workers` | Number of concurrent workers | 5 |
| `--output` | Output directory | ./data |
| `--exclude` | Domains to exclude | None |
| `--verbose` | Enable verbose logging | false |
| `--port` | API server port (serve mode) | 8080 |

## 🏗️ Architecture

### System Design

```
┌─────────────────────────────────────────────────────┐
│                   Main Controller                    │
└────────────────┬────────────────────────────────────┘
                 │
    ┌────────────┴────────────┬──────────────┐
    │                         │              │
┌───▼────┐              ┌─────▼─────┐   ┌───▼────┐
│Frontier│◄────────────►│  Workers  │   │Storage │
│        │   URLs       │  (Pool)   │   │        │
│ Queue  │              │           │   │ System │
└────────┘              └─────┬─────┘   └────────┘
                              │
                        ┌─────▼──────┐
                        │  Parsers   │
                        │ (HTML/URL) │
                        └────────────┘
```

### Key Components

- **Frontier**: Manages URL queue with deduplication
- **Worker Pool**: Concurrent HTTP fetchers with rate limiting
- **Parser**: Extracts links and content from HTML
- **Storage**: Persists crawled content to disk
- **Processors**: Extensible pipeline for custom processing

### Worker Pool Pattern

```go
// Configurable number of workers
workers := 10

// Each worker processes URLs concurrently
for i := 0; i < workers; i++ {
    go worker.Start()
}

// Politeness delay per domain
time.Sleep(2 * time.Second)
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

| Package | Coverage | Tests |
|---------|----------|-------|
| `crawler` | 78% | 13 tests |
| `frontier` | 72% | 9 tests |
| **Total** | **75%** | **22 tests** |

## 🐳 Docker

### Build Image

```bash
# Build production image
docker build -t web-crawler:latest .

# Build without cache
docker build --no-cache -t web-crawler:latest .
```

### Image Details

- **Base Image**: Alpine Linux (minimal, secure)
- **Size**: ~15-20 MB (optimized with multi-stage build)
- **User**: Non-root user (security best practice)
- **Architecture**: Supports amd64 and arm64

### Multi-Stage Build

```dockerfile
# Stage 1: Build (golang:alpine)
- Download dependencies
- Run tests
- Compile binary

# Stage 2: Runtime (alpine:latest)
- Minimal image
- Copy binary only
- Non-root user
```

## 🔄 CI/CD Pipeline

Automated pipeline runs on every commit:

```
Push to GitHub
    ↓
┌───────────────┐
│  Run Tests    │  ← go test -v ./crawler
└───────┬───────┘
        ↓
┌───────────────┐
│  Build Binary │  ← go build
└───────┬───────┘
        ↓
┌───────────────┐
│ Build Docker  │  ← docker build
└───────┬───────┘
        ↓
┌───────────────┐
│ Verify Image  │  ← docker run --help
└───────────────┘
```

**Pipeline Time**: ~2 minutes ⚡

## 📊 Performance

### Benchmarks

| Metric | Value |
|--------|-------|
| Avg Response Time | ~250ms |
| URLs/second | ~40-50 (depends on workers) |
| Memory Usage | ~50-100 MB |
| CPU Usage | ~1-2 cores (10 workers) |

### Optimization Features

- ✅ Connection pooling
- ✅ Concurrent processing
- ✅ Efficient deduplication (map-based)
- ✅ Minimal memory footprint
- ✅ Rate limiting per domain

## 🛠️ Development

### Project Structure

```
web_crawler/
├── crawler/              # Core crawler logic
│   ├── crawler.go       # Main crawler orchestration
│   ├── worker.go        # Worker pool implementation
│   ├── processor.go     # Content processors
│   ├── queue.go         # URL queue management
│   └── *_test.go        # Test files
├── frontier/            # URL frontier (queue + dedup)
│   ├── frontier.go
│   └── frontier_test.go
├── parser/              # HTML parsing & link extraction
│   └── parser.go
├── storage/             # Content storage system
│   └── storage.go
├── .github/
│   └── workflows/       # CI/CD pipelines
│       └── ci-simple.yml
├── Dockerfile           # Docker image definition
├── docker-compose.yml   # Multi-container orchestration
├── go.mod               # Go dependencies
└── main.go              # CLI entry point
```

### Tech Stack

- **Language**: Go 1.24
- **CLI Framework**: Cobra
- **Testing**: Go testing framework
- **Containerization**: Docker
- **CI/CD**: GitHub Actions
- **Architecture**: Worker pool, channel-based concurrency

### Adding Custom Processors

```go
// Implement the Processor interface
type MyProcessor struct {}

func (p *MyProcessor) Process(result *CrawlResult) error {
    // Custom processing logic
    return nil
}

// Add to crawler
crawler.AddProcessor(&MyProcessor{})
```

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Quality

- Write tests for new features
- Maintain >70% code coverage
- Follow Go best practices
- Run `go fmt` before committing
- Ensure CI pipeline passes

## 📝 Future Enhancements

- [ ] **Web Dashboard** - React-based UI for crawler management
- [ ] **Database Storage** - PostgreSQL/MongoDB support
- [ ] **Distributed Crawling** - Multi-node coordination
- [ ] **Robots.txt Support** - Respect crawl rules
- [ ] **Webhook Notifications** - Real-time crawl updates
- [ ] **Content Analysis** - NLP and sentiment analysis
- [ ] **Export Formats** - CSV, JSON, XML output
- [ ] **Metrics Dashboard** - Prometheus/Grafana integration

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- CLI powered by [Cobra](https://github.com/spf13/cobra)
- Inspired by best practices in web crawling

## 📧 Contact

**Fardin Rahman** - [@voidpntr](https://x.com/voidpntr) - fardinrahman647@gmail.com

Project Link: [https://github.com/Fardin-E/web_crawler](https://github.com/Fardin-E/web_crawler)

---

<div align="center">
  
**⭐ Star this repository if you find it helpful!**

Made with ❤️ and Go

</div>