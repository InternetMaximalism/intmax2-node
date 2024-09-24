-- +migrate Up

ALTER TABLE backup_transfers DROP COLUMN sender_last_balance_proof_body;
ALTER TABLE backup_transfers DROP COLUMN sender_balance_transition_proof_body;

CREATE TABLE backup_sender_proofs (
    id uuid not null default uuid_generate_v4(),
    enough_balance_proof_body_hash varchar(66) NOT NULL,
    last_balance_proof_body bytea,
    balance_transition_proof_body bytea,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

-- ALTER TABLE backup_transactions ADD COLUMN sender_enough_balance_proof_body_hash varchar(66);

-- +migrate Down

-- ALTER TABLE backup_transactions DROP COLUMN sender_enough_balance_proof_body_hash;

DROP TABLE backup_sender_proofs;

ALTER TABLE backup_transfers ADD COLUMN sender_last_balance_proof_body bytea;
ALTER TABLE backup_transfers ADD COLUMN sender_balance_transition_proof_body bytea;
