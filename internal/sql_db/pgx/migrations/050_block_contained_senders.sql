-- +migrate Up

DROP TABLE block_contained_senders;

-- +migrate Down

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE block_contained_senders (
    id          uuid not null default uuid_generate_v4(),
    block_hash  varchar(64) not null,
    sender      varchar(64) not null,
    PRIMARY KEY (id)
);
