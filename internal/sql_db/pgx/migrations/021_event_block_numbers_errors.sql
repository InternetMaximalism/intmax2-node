-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE event_block_numbers_errors (
    id           uuid not null default uuid_generate_v4(),
    event_name   varchar(255) not null,
    block_number numeric not null,
    options      bytea not null,
    errors       jsonb not null default '{}'::jsonb,
    created_at   timestamptz not null default now(),
    updated_at   timestamptz not null default now(),
    PRIMARY KEY (id),
    CONSTRAINT check_event_block_numbers_errors_bn_positive CHECK (block_number >= 0)
);

CREATE UNIQUE INDEX idx_unique_event_block_numbers_errors_event_name_bn
    ON event_block_numbers_errors(event_name, block_number);

CREATE INDEX
    IF NOT EXISTS idx_event_block_numbers_errors_created_at_id
    ON event_block_numbers_errors(created_at, id);

-- +migrate Down

DROP TABLE event_block_numbers_errors;
