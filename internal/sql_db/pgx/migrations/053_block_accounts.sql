-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SEQUENCE IF NOT EXISTS block_accounts_account_id_seq;

CREATE TABLE block_accounts (
    id         uuid not null default uuid_generate_v4(),
    account_id numeric NOT NULL DEFAULT nextval('block_accounts_account_id_seq'),
    sender_id  uuid not null references block_senders(id),
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX idx_block_accounts_sender_id ON block_accounts(sender_id);

ALTER SEQUENCE block_accounts_account_id_seq RESTART WITH 2;

-- +migrate Down

DROP TABLE block_accounts;
