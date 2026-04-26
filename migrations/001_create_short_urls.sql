CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS short_urls (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    original_url text NOT NULL,
    short_code varchar(20) NOT NULL UNIQUE,
    click_count integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    expires_at timestamptz,
    is_active boolean NOT NULL DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_short_urls_short_code ON short_urls (short_code);
CREATE INDEX IF NOT EXISTS idx_short_urls_is_active ON short_urls (is_active);
CREATE INDEX IF NOT EXISTS idx_short_urls_expires_at ON short_urls (expires_at);
