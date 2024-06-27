-- +migrate Up

CREATE SEQUENCE IF NOT EXISTS signatures_signature_id_seq;

CREATE TABLE signatures (
    signature_id      numeric NOT NULL DEFAULT nextval('signatures_signature_id_seq'),
    signature         varchar(255) not null,
    proposal_block_id numeric not null references blocks(proposal_block_id),
    created_at        timestamptz not null default now(),
    PRIMARY KEY (signature_id)
);

CREATE INDEX
    IF NOT EXISTS idx_signatures_proposal_block_id
    ON signatures(proposal_block_id);

-- +migrate Down

DROP TABLE signatures;
