-- +migrate Up

CREATE SEQUENCE IF NOT EXISTS block_contents_block_number_seq;

CREATE TABLE block_contained_senders (
    id          uuid not null default uuid_generate_v4(),
    block_hash  varchar(64) not null,
    sender      varchar(64) not null,
    PRIMARY KEY (id)
);

CREATE TABLE block_contents (
    id                    uuid not null default uuid_generate_v4(),
    block_number          bigint not null default nextval('block_contents_block_number_seq'),
    block_hash            varchar(64) not null,
    prev_block_hash       varchar(64) not null,
    deposit_root          varchar(64) not null,
    is_registration_block boolean not null,
    senders               jsonb not null,
    tx_tree_root          varchar(64) not null,
    aggregated_public_key varchar(128) not null,
    aggregated_signature  varchar(256) not null,
    message_point         varchar(256) not null,
    created_at            timestamptz not null default now(),
    PRIMARY KEY (id),
    CONSTRAINT check_event_block_numbers_errors_bn_positive CHECK (block_number >= 0)
);

-- +migrate Down

DROP TABLE block_contents;

DROP TABLE block_contained_senders;

DROP SEQUENCE block_contents_block_number_seq;