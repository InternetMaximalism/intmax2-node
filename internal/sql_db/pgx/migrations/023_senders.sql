-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE senders (
    id         uuid not null default uuid_generate_v4(),
    address    varchar not null unique,
    public_key varchar not null unique,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

-- +migrate Down

DROP TABLE senders;
