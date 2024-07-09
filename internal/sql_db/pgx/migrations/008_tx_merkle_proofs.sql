-- +migrate Up

DROP TABLE tx_merkle_proofs;
DROP SEQUENCE IF EXISTS tx_merkle_proofs_id_seq;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tx_merkle_proofs (
    id                uuid not null default uuid_generate_v4(),
    sender_public_key varchar(255) not null,
    tx_hash           varchar(255) not null,
    tx_tree_index     numeric not null,
    tx_merkle_proof   jsonb not null,
    created_at        timestamptz not null default now(),
    PRIMARY KEY (id),
    unique (sender_public_key, tx_hash)
);

CREATE INDEX
    IF NOT EXISTS idx_tx_merkle_proofs_tx_hash
    ON tx_merkle_proofs(tx_hash);

CREATE INDEX
    IF NOT EXISTS idx_tx_merkle_proofs_created_at_id
    ON tx_merkle_proofs(created_at, id);

-- +migrate Down

DROP TABLE tx_merkle_proofs;

CREATE SEQUENCE IF NOT EXISTS tx_merkle_proofs_id_seq;

CREATE TABLE tx_merkle_proofs (
    id                numeric NOT NULL DEFAULT nextval('tx_merkle_proofs_id_seq'),
    sender_public_key varchar(255) not null,
    tx_hash           varchar(255) not null references transactions(tx_hash),
    tx_tree_index     numeric not null,
    tx_merkle_proof   bytea not null,
    created_at        timestamptz not null default now(),
    PRIMARY KEY (id),
    unique (sender_public_key, tx_hash)
);

DROP SEQUENCE IF EXISTS tx_merkle_proofs_tx_tree_index_seq;

CREATE INDEX
    IF NOT EXISTS idx_tx_merkle_proofs_tx_hash
    ON tx_merkle_proofs(tx_hash);
