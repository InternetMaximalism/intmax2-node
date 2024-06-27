-- +migrate Up

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

CREATE INDEX
    IF NOT EXISTS idx_tx_merkle_proofs_tx_hash
    ON tx_merkle_proofs(tx_hash);

-- +migrate Down

DROP TABLE tx_merkle_proofs;
