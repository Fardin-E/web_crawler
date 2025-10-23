# Makefile for Web Crawler

.PHONY: help build test docker-build docker-run docker-push clean

# Variables
APP_NAME=web-crawler
VERSION?=latest
DOCKER_IMAGE=${APP_NAME}:${VERSION}
DOCKER_REGISTRY?=

# Help command
help:
	@echo "Web Crawler - Available Commands:"
	@echo ""
	@echo "  make build          - Build the Go binary"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run crawler in Docker"
	@echo "  make docker-api     - Run API server in Docker"
	@echo "  make docker-push    - Push image to registry"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make all            - Test, build, and dockerize"

# Build Go binary
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME).exe .
	@echo "Build complete: $(APP_NAME).exe"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./crawler ./frontier

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -cover ./crawler ./frontier
	go test -coverprofile=coverage.out ./crawler ./frontier
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Docker build
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE)"
	docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built successfully"

# Run crawler in Docker
docker-run:
	@echo "Running crawler in Docker..."
	docker run --rm \
		-v $$(pwd)/data:/home/crawler/data \
		$(DOCKER_IMAGE) \
		crawl --url https://golang.org --workers 5 --verbose

# Run API server in Docker
docker-api:
	@echo "Starting API server in Docker..."
	docker run --rm \
		-p 8080:8080 \
		$(DOCKER_IMAGE) \
		serve --port 8080 --host 0.0.0.0

# Push to Docker registry
docker-push:
	@if [ -z "$(DOCKER_REGISTRY)" ]; then \
		echo "Error: DOCKER_REGISTRY not set"; \
		echo "Usage: make docker-push DOCKER_REGISTRY=your-registry.com"; \
		exit 1; \
	fi
	docker tag $(DOCKER_IMAGE) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME) $(APP_NAME).exe
	rm -f coverage.out coverage.html
	rm -rf data/
	@echo "Clean complete"

# Run everything
all: test build docker-build
	@echo ""
	@echo "✅ All tasks completed successfully!"
	@echo "Docker image: $(DOCKER_IMAGE)"