CREATE TABLE IF NOT EXISTS aegis_vault_keeper.auth_users (
    id UUID PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    crypto_key BYTEA NOT NULL
);