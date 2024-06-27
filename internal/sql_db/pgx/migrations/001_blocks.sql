-- +migrate Up

CREATE SEQUENCE IF NOT EXISTS blocks_proposal_block_id_seq;

DROP TYPE IF EXISTS blocks_status;
CREATE TYPE blocks_status AS ENUM ('pending', 'committed', 'confirmed', 'failed');

CREATE TABLE blocks (
    proposal_block_id     numeric NOT NULL DEFAULT nextval('blocks_proposal_block_id_seq'),
    builder_public_key    varchar(255) not null,
    tx_root               varchar(255) not null,
    block_hash            varchar(255) not null,
    aggregated_signature  varchar(255) not null,
    aggregated_public_key varchar(255) not null,
    status                blocks_status not null default 'pending'::blocks_status,
    created_at            timestamptz not null default now(),
    posted_at             timestamptz,
    PRIMARY KEY (proposal_block_id)
);

-- +migrate Down

DROP TABLE blocks;
DROP TYPE IF EXISTS blocks_status;
