-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TYPE IF EXISTS transactions_status;
CREATE TYPE transactions_status AS ENUM ('pending', 'committed', 'confirmed', 'failed');

CREATE TABLE transactions (
    tx_id             uuid not null default uuid_generate_v4(),
    tx_hash           varchar(255),
    sender_public_key varchar(255) not null,
    signature_id      numeric not null references signatures(signature_id),
    status            transactions_status not null default 'pending'::transactions_status,
    created_at        timestamptz not null default now(),
    PRIMARY KEY (tx_id)
);

CREATE UNIQUE INDEX
    IF NOT EXISTS idx_unique_transactions_tx_hash
    ON transactions(tx_hash);

CREATE INDEX
    IF NOT EXISTS idx_transactions_created_at_tx_id
    ON transactions(created_at, tx_id);

CREATE INDEX
    IF NOT EXISTS idx_transactions_signature_id
    ON transactions(signature_id);

CREATE INDEX
    IF NOT EXISTS idx_transactions_status
    ON transactions(status);

-- +migrate Down

DROP TABLE transactions;
DROP TYPE IF EXISTS transactions_status;
