-- +migrate Up

ALTER TABLE backup_transfers ADD COLUMN sender_last_balance_proof_body bytea;
ALTER TABLE backup_transfers ADD COLUMN sender_balance_transition_proof_body bytea;

ALTER TABLE backup_transactions ADD COLUMN encoding_version integer NOT NULL DEFAULT 0;

-- +migrate Down

ALTER TABLE backup_transfers DROP COLUMN sender_last_balance_proof_body;
ALTER TABLE backup_transfers DROP COLUMN sender_balance_transition_proof_body;

ALTER TABLE backup_transactions DROP COLUMN encoding_version;
