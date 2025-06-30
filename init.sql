-- Create a schema
CREATE SCHEMA IF NOT EXISTS crawler_schema;

-- Create a table within that schema
CREATE TABLE crawler_schema.pages (
    id BIGSERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    status_code SMALLINT NOT NULL,
    content_type TEXT,
    title TEXT,
    meta_description TEXT,
    content_length INTEGER,
    response_time_ms INTEGER,
    out_links TEXT[], -- Good as TEXT[]
    is_error BOOLEAN DEFAULT FALSE,
    raw_html TEXT, -- New dedicated column for raw HTML
    paragraphs TEXT[], -- New dedicated column for parsed paragraphs (using PG's native array type)
    -- You could still keep a generic 'parsed_data JSONB' column if you have other misc. parsed info
    -- that doesn't warrant its own dedicated column.
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);