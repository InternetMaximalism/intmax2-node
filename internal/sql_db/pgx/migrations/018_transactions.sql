-- +migrate Up

ALTER TABLE tx_merkle_proofs DROP COLUMN tx_id;

DROP TABLE transactions;
DROP TYPE IF EXISTS transactions_status;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TABLE signatures;
CREATE TABLE signatures (
  signature_id      uuid not null default uuid_generate_v4(),
  signature         varchar not null,
  proposal_block_id uuid not null references blocks(proposal_block_id),
  created_at        timestamptz not null default now(),
  PRIMARY KEY (signature_id)
);

CREATE INDEX
    IF NOT EXISTS idx_signatures_proposal_block_id
    ON signatures(proposal_block_id);

ALTER TABLE blocks ALTER COLUMN aggregated_signature TYPE VARCHAR USING aggregated_signature::varchar;
ALTER TABLE blocks ALTER COLUMN aggregated_public_key TYPE VARCHAR USING aggregated_public_key::varchar;
ALTER TABLE blocks ALTER COLUMN block_hash DROP NOT NULL;
ALTER TABLE blocks ALTER COLUMN status DROP NOT NULL;
ALTER TABLE blocks ADD COLUMN sender_type int not null;
ALTER TABLE blocks ADD COLUMN options bytea;

ALTER TABLE tx_merkle_proofs ADD COLUMN signature_id uuid references signatures;
ALTER TABLE tx_merkle_proofs ADD COLUMN tx_tree_root varchar(255) not null;
ALTER TABLE tx_merkle_proofs ADD COLUMN proposal_block_id uuid not null references blocks(proposal_block_id);

-- +migrate Down

ALTER TABLE tx_merkle_proofs DROP COLUMN signature_id;
ALTER TABLE tx_merkle_proofs DROP COLUMN proposal_block_id;
ALTER TABLE tx_merkle_proofs DROP COLUMN tx_tree_root;

DROP TABLE signatures;
CREATE SEQUENCE IF NOT EXISTS signatures_signature_id_seq;
CREATE TABLE signatures (
  signature_id numeric NOT NULL DEFAULT nextval('signatures_signature_id_seq'),
  signature    varchar(255) not null,
  proposal_block_id uuid references blocks(proposal_block_id),
  created_at   timestamptz not null default now(),
  PRIMARY KEY (signature_id)
);

CREATE INDEX
    IF NOT EXISTS idx_signatures_proposal_block_id
    ON signatures(proposal_block_id);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE transactions
(
    tx_id             uuid default uuid_generate_v4() not null primary key,
    tx_hash           varchar(255) not null,
    sender_public_key varchar(255) not null,
    signature_id      numeric not null references signatures,
    created_at        timestamp with time zone default now() not null,
    tx_tree_info      jsonb default '{}'::jsonb not null,
    status            integer
);

CREATE INDEX idx_transactions_created_at_tx_id ON transactions (created_at, tx_id);
CREATE INDEX idx_transactions_signature_id ON transactions (signature_id);
CREATE UNIQUE INDEX idx_unique_transactions_tx_hash ON transactions (tx_hash);

ALTER TABLE tx_merkle_proofs ADD COLUMN tx_id uuid not null references transactions(tx_id);

ALTER TABLE blocks ALTER COLUMN aggregated_signature TYPE varchar(255) using aggregated_signature::varchar(255);
ALTER TABLE blocks ALTER COLUMN aggregated_public_key TYPE varchar(255) using aggregated_public_key::varchar(255);
ALTER TABLE blocks ALTER COLUMN block_hash SET NOT NULL;
ALTER TABLE blocks ALTER COLUMN status SET NOT NULL;
ALTER TABLE blocks DROP COLUMN sender_type;
ALTER TABLE blocks DROP COLUMN options;

