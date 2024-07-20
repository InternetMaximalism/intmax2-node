-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

ALTER TABLE tx_merkle_proofs ADD COLUMN tx_id uuid not null references transactions(tx_id);

-- +migrate Down

ALTER TABLE tx_merkle_proofs DROP COLUMN tx_id;
