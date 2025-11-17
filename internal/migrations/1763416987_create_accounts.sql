-- Creates the core accounts table used by the storage layer.
-- Run this against the local Postgres instance (see docker-compose.local.yml).

CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    balance REAL NOT NULL DEFAULT 0
);

