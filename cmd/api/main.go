package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

type apiDB struct {
	pool *pgxpool.Pool
}

func (a *apiDB) ConnectDB() error {
	// Load .env file for local development (won't affect Docker environments)
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file (this is normal if running in Docker): %v", err)
	}

	// Retrieve database credentials from environment variables
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPassword := os.Getenv("DATABASE_PASSWORD")
	dbName := os.Getenv("DATABASE_NAME")

	// Basic validation to ensure critical environment variables are set
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		return fmt.Errorf("one or more essential database environment variables are not set. Check DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME")
	}

	// Construct the DSN (Data Source Name) URL using environment variables
	dsn := url.URL{
		Scheme: "postgres",
		Host:   fmt.Sprintf("%s:%s", dbHost, dbPort),
		User:   url.UserPassword(dbUser, dbPassword),
		Path:   fmt.Sprintf("/%s", dbName), // Path is the database name
		// You might add RawQuery for sslmode, e.g., RawQuery: "sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn.String())
	if err != nil {
		return fmt.Errorf("DB connection pool creation failed: %w", err)
	}

	// Ping the database to verify the connection
	err = pool.Ping(context.Background())
	if err != nil {
		pool.Close() // Close the pool if ping fails
		return fmt.Errorf("DB connection verification failed (ping error): %w", err)
	}

	a.pool = pool // Assign the created pool to the struct field
	log.Println("Connected to DB successfully.")
	return nil
}

func (a *apiDB) CloseDB() {
	if a.pool != nil {
		a.pool.Close()
		fmt.Println("Database connection closed")
	}
}

// PageData represents the structure of the data you want to send to the frontend
type PageData struct {
	ID        int    `json:"id"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Snippet   string `json:"snippet"`
	CrawledAt string `json:"crawledAt"`
}

// Now, the global variable that handlers will use needs to be of type *apiDB
var appDB *apiDB

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
}
