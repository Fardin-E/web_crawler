-- Create a schema
CREATE SCHEMA IF NOT EXISTS crawler_schema;

-- Create a table within that schema
CREATE TABLE IF NOT EXISTS crawler_schema.pages (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,                   -- Page URL
    status_code SMALLINT,                       -- HTTP status code (e.g., 200, 404)
    content_type TEXT,                          -- e.g., "text/html", "application/json"
    html_content JSONB,                         -- extracted HTML infos
    title TEXT,                                 -- Page <title>
    meta_description TEXT,                      -- From <meta name="description">
    content_length INTEGER,                     -- Length of the body content
    fetched_at TIMESTAMP WITH TIME ZONE DEFAULT now(), -- Crawl timestamp
    response_time_ms INTEGER,                   -- Response time in ms
    out_links TEXT[],                           -- Array of links found on page
    is_error BOOLEAN DEFAULT false 
);