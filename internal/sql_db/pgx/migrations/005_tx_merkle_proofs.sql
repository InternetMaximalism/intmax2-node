-- +migrate Up

CREATE TABLE tx_merkle_proofs (
    id                bigserial primary key,
    sender_public_key varchar(255) not null,
    tx_hash           varchar(255) not null references transactions(tx_hash),
    tx_tree_index     bigint not null,
    tx_merkle_proof   bytea not null,
    created_at        timestamptz not null default now(),
    unique (sender_public_key, tx_hash)
);

-- +migrate Down

DROP TABLE tx_merkle_proofs;
