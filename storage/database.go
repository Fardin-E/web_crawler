package storage

/*

CREATE TABLE crawler_schema.pages (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,                   -- Page URL
    status_code SMALLINT,                       -- HTTP status code (e.g., 200, 404)
    content_type TEXT,                          -- e.g., "text/html", "application/json"
    html_content JSONB,                          -- Full HTML of the page
    title TEXT,                                 -- Page <title>
    meta_description TEXT,                      -- From <meta name="description">
    content_length INTEGER,                     -- Length of the body content
    fetched_at TIMESTAMP WITH TIME ZONE DEFAULT now(), -- Crawl timestamp
    response_time_ms INTEGER,                   -- Response time in ms
    out_links TEXT[],                           -- Array of links found on page
    is_error BOOLEAN DEFAULT false              -- True if crawl failed
);
*/

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type ConnDB struct {
	db *pgxpool.Pool
}

func (c *ConnDB) ConnectDB() error {
	dsn := url.URL{
		Scheme: "postgres",
		Host:   "localhost:5432",
		User:   url.UserPassword("freid", "password"),
		Path:   "/crawler",
	}

	pool, err := pgxpool.New(context.Background(), dsn.String())
	if err != nil {
		fmt.Println("DB connection error:", err)
		return fmt.Errorf("DB connection failed: %w", err)
	}

	c.db = pool
	fmt.Println("Connected to DB")
	return nil
}

func (c *ConnDB) CloseDB() {
	if c.db != nil {
		c.db.Close()
		fmt.Println("Database connection closed")
	}
}

func (c *ConnDB) InsertPage(ctx context.Context, url string, content_type string, body []byte) error {
	sql := `INSERT INTO crawler_schema.pages (url, content_type, html_content) VALUES ($1, $2, $3)`
	_, err := c.db.Exec(ctx, sql, url, content_type, body)
	return err
}
