-- Creates the core accounts table used by the storage layer.
-- Run this against the local Postgres instance (see docker-compose.local.yml).

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    balance NUMERIC(23, 5) NOT NULL DEFAULT 0
);

