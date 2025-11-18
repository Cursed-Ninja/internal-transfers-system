-- Creates the transactions table used to log all internal transfers.
-- Both source_account_id and destination_account_id reference accounts.id
-- Deleting an account with existing transactions is restricted
-- created_at records the timestamp of the transaction
-- Run this against the local Postgres instance (see docker-compose.local.yml).

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    source_account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    destination_account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
    amount NUMERIC(23, 5) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);