-- +migrate Up

DROP TABLE deposits;

DROP TABLE event_block_numbers_validity_prover;

DROP TABLE block_contents;

CREATE TABLE block_contents (
    id                    uuid not null default uuid_generate_v4(),
    block_number          bigint not null default nextval('block_contents_block_number_seq'),
    block_hash            varchar(64) not null,
    prev_block_hash       varchar(64) not null,
    deposit_root          varchar(64) not null,
    signature_hash        varchar(64) not null,
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

CREATE TABLE event_block_numbers_validity_prover (
    id UUID not null default uuid_generate_v4(),
    event_name varchar(255) not null unique,
    last_processed_block_number int not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id),
    CONSTRAINT check_last_processed_block_number_validity_prover_positive CHECK (last_processed_block_number >= 0)
);

CREATE INDEX idx_event_block_numbers_validity_prover_event_name ON event_block_numbers_validity_prover(event_name);

CREATE TABLE deposits (
    id                    uuid not null default uuid_generate_v4(),
    deposit_id            bigint not null,
    deposit_hash          varchar(255) not null,
    recipient_salt_hash   varchar(255) not null,
    token_index           bigint not null,
    amount                varchar(255) not null,
    deposit_index         bigint,
    created_at            timestamptz not null default now(),
    PRIMARY KEY (id),
    UNIQUE (deposit_id)
);

-- +migrate Down

DROP TABLE deposits;

DROP TABLE event_block_numbers_validity_prover;

DROP TABLE block_contents;

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

CREATE TABLE event_block_numbers_validity_prover (
    id UUID not null default uuid_generate_v4(),
    event_name varchar(255) not null unique,
    last_processed_block_number int not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id),
    CONSTRAINT check_last_processed_block_number_validity_prover_positive CHECK (last_processed_block_number >= 0)
);

CREATE INDEX idx_event_block_numbers_validity_prover_event_name ON event_block_numbers_validity_prover(event_name);

CREATE TABLE deposits (
    id                    uuid not null default uuid_generate_v4(),
    deposit_id            bigint not null,
    deposit_hash          varchar(255) not null,
    recipient_salt_hash   varchar(255) not null,
    token_index           bigint not null,
    amount                varchar(255) not null,
    deposit_index         bigint,
    created_at            timestamptz not null default now(),
    PRIMARY KEY (id),
    UNIQUE (deposit_id)
);
