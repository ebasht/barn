-- +goose Up
CREATE TABLE mcp_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    instance_id TEXT NOT NULL UNIQUE,
    instance_name TEXT NOT NULL,
    token_hash BYTEA CHECK (token_hash IS NULL OR octet_length(token_hash) = 32),
    allow_writes BOOLEAN NOT NULL DEFAULT FALSE,
    key_created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS mcp_settings;
