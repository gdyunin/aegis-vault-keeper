CREATE TABLE IF NOT EXISTS aegis_vault_keeper.credentials
(
    id          UUID      PRIMARY KEY,
    user_id     UUID      NOT NULL,
    login       BYTEA     NOT NULL,
    password    BYTEA     NOT NULL,
    description BYTEA     NOT NULL,
    updated_at  TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS aegis_vault_keeper.notes
(
    id          UUID      PRIMARY KEY,
    user_id     UUID      NOT NULL,
    note        BYTEA     NOT NULL,
    description BYTEA     NOT NULL,
    updated_at  TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS aegis_vault_keeper.bank_cards
(
    id           UUID      PRIMARY KEY,
    user_id      UUID      NOT NULL,
    card_number  BYTEA     NOT NULL,
    card_holder  BYTEA     NOT NULL,
    expiry_month BYTEA     NOT NULL,
    expiry_year  BYTEA     NOT NULL,
    cvv          BYTEA     NOT NULL,
    description  BYTEA     NOT NULL,
    updated_at   TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS aegis_vault_keeper.files (
    id          UUID PRIMARY KEY,
    user_id     UUID NOT NULL,
    description BYTEA NOT NULL,
    updated_at  TIMESTAMP NOT NULL,
    storage_key BYTEA NOT NULL,
    hash_sum    BYTEA NOT NULL
);