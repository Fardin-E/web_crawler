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

// InsertPage now takes all the necessary data to populate the columns
func (c *ConnDB) InsertPage(
	ctx context.Context,
	url string,
	statusCode int,
	contentType string,
	title string,
	metaDescription string,
	contentLength int,
	responseTimeMs int,
	outLinks []string, // PostgreSQL TEXT[] array
	isError bool,
	rawHtml string, // ADDED: For the full raw HTML
	paragraphs []string, // ADDED: For the extracted paragraphs (TEXT[] in DB)
) error {
	sql := `
        INSERT INTO crawler_schema.pages (
            url, status_code, content_type,
            title, meta_description, content_length,
            response_time_ms, out_links, is_error,
            raw_html, paragraphs -- ADDED: new columns here
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11 -- Updated number of placeholders
        )
        ON CONFLICT (url) DO UPDATE SET
            status_code = EXCLUDED.status_code,
            content_type = EXCLUDED.content_type,
            -- html_content = EXCLUDED.html_content, -- REMOVED: no longer using this column
            title = EXCLUDED.title,
            meta_description = EXCLUDED.meta_description,
            content_length = EXCLUDED.content_length,
            updated_at = NOW(), -- Assumed you have an updated_at column for the timestamp
            response_time_ms = EXCLUDED.response_time_ms,
            out_links = EXCLUDED.out_links,
            is_error = EXCLUDED.is_error,
            raw_html = EXCLUDED.raw_html,     -- ADDED: update raw_html on conflict
            paragraphs = EXCLUDED.paragraphs; -- ADDED: update paragraphs on conflict
    `
	// The number of parameters for Exec must match the number of placeholders in the SQL.
	_, err := c.db.Exec(
		ctx,
		sql,
		url,
		statusCode,
		contentType,
		title,
		metaDescription,
		contentLength,
		responseTimeMs,
		outLinks,
		isError,
		rawHtml,
		paragraphs,
	)
	return err
}
