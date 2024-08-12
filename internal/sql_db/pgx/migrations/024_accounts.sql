-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE SEQUENCE IF NOT EXISTS accounts_account_id_seq;

CREATE TABLE accounts (
    id         uuid not null default uuid_generate_v4(),
    account_id numeric NOT NULL DEFAULT nextval('accounts_account_id_seq'),
    sender_id  uuid not null references senders(id),
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE UNIQUE INDEX idx_accounts_sender_id ON accounts(sender_id);

-- +migrate Down

DROP TABLE accounts;
