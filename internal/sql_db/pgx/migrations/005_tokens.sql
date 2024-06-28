-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tokens (
    id            uuid not null default uuid_generate_v4(),
    token_index   varchar(255) not null,
    token_address varchar(255),
    token_id      numeric not null,
    created_at    timestamptz not null default now(),
    PRIMARY KEY (id),
    unique (token_index)
);

CREATE INDEX
    IF NOT EXISTS idx_tokens_created_at_id
    ON tokens(created_at, id);

-- +migrate Down

DROP TABLE tokens;
