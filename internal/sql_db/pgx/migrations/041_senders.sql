-- +migrate Up

ALTER TABLE deposits
DROP CONSTRAINT IF EXISTS unique_deposits_deposit_index;

ALTER TABLE block_contents
DROP CONSTRAINT IF EXISTS unique_block_contents_block_number;

DROP TABLE block_validity_proofs;

DROP TABLE deposits;

DROP TABLE event_block_numbers_validity_prover;

DROP TABLE block_contained_senders;

DROP TABLE block_contents;

DROP TABLE accounts;

DROP TABLE senders;

CREATE TABLE senders (
    id         uuid not null default uuid_generate_v4(),
    address    varchar(64) not null,
    public_key varchar(128) not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id),
    UNIQUE (address)
);


CREATE TABLE accounts (
    id         uuid not null default uuid_generate_v4(),
    account_id numeric NOT NULL DEFAULT nextval('accounts_account_id_seq'),
    sender_id  uuid not null references senders(id),
    created_at timestamptz not null default now(),
    PRIMARY KEY (id),
    UNIQUE (account_id)
);

CREATE UNIQUE INDEX idx_accounts_sender_id ON accounts(sender_id);

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
    UNIQUE (block_number),
    CONSTRAINT check_event_block_numbers_errors_bn_positive CHECK (block_number >= 0)
);

CREATE TABLE block_contained_senders (
    id           uuid not null default uuid_generate_v4(),
    block_number bigint not null,
    sender_id    uuid not null,
    created_at   timestamptz not null default now(),
    PRIMARY KEY (id),
    FOREIGN KEY (sender_id) REFERENCES senders(id) ON DELETE CASCADE,
    UNIQUE (block_number, sender_id)
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
    UNIQUE (deposit_id),
    UNIQUE (deposit_index)
);

CREATE TABLE block_validity_proofs (
    id               uuid not null default uuid_generate_v4(),
    block_content_id uuid not null,
    validity_proof   bytea not null,
    PRIMARY KEY (id),
    FOREIGN KEY (block_content_id) REFERENCES block_contents(id) ON DELETE CASCADE,
    UNIQUE (block_content_id)
);

-- ALTER TABLE block_contained_senders
-- DROP COLUMN sender;

-- ALTER TABLE block_contained_senders
-- ADD COLUMN sender_id uuid not null;

-- ALTER TABLE block_contained_senders
-- ADD COLUMN created_at timestamptz not null default now();

-- ALTER TABLE block_contained_senders
-- ADD CONSTRAINT fk_block_contained_senders_sender_id
-- FOREIGN KEY (sender_id) REFERENCES senders(id);

WITH tmp_senders AS (
    INSERT INTO senders (address, public_key)
    VALUES ('0000000000000000000000000000000000000000000000000000000000000001', '0000000000000000000000000000000000000000000000000000000000000001')
    ON CONFLICT (address) DO NOTHING
    RETURNING id
)
INSERT INTO accounts (account_id, sender_id)
SELECT 1, COALESCE(
    (SELECT id FROM tmp_senders),
    (SELECT id FROM senders WHERE address = '0000000000000000000000000000000000000000000000000000000000000001')
) ON CONFLICT (sender_id) DO NOTHING;

ALTER SEQUENCE accounts_account_id_seq RESTART WITH 2;

-- +migrate Down

-- ALTER TABLE block_contained_senders
-- DROP CONSTRAINT fk_block_contained_senders_sender_id;

ALTER TABLE block_contained_senders
DROP COLUMN created_at;

ALTER TABLE block_contained_senders
DROP COLUMN sender_id;

ALTER TABLE block_contained_senders
ADD COLUMN sender varchar(64) not null;

ALTER TABLE block_contained_senders
DROP COLUMN block_number;

ALTER TABLE block_contained_senders
ADD COLUMN block_hash varchar(64) not null;
