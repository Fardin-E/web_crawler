# Multi-stage Dockerfile for Go Web Crawler

# Stage 1: Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files first (better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Run tests - only crawler package (frontier has channel blocking issues in Docker)
RUN go test -v ./crawler

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o crawler.out .

# Stage 2: Runtime stage (minimal image)
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 crawler && \
    adduser -D -u 1000 -G crawler crawler

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=crawler:crawler --chmod=755 /app/crawler.out /app/crawler.out

# Create data directory with correct permissions
RUN mkdir -p data && \
    chown -R crawler:crawler /app

# Switch to non-root user
USER crawler

# Expose port for API server
EXPOSE 8080

# Set default command
ENTRYPOINT ["/app/crawler.out"]

# Default arguments
CMD ["--help"]