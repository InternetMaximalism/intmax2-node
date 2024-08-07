-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE balances (
    id           uuid not null default uuid_generate_v4(),
    user_address varchar(255) not null,
    token_index  varchar(255) not null,
    balance      varchar(255) not null,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    PRIMARY KEY (id),
    UNIQUE (user_address, token_index),
    FOREIGN KEY (token_index) REFERENCES tokens(token_index)
);

-- +migrate Down

DROP TABLE balances;
