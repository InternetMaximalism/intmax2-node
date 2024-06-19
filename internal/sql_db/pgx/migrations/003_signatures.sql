-- +migrate Up

CREATE TABLE signatures (
    signature_id      bigserial primary key,
    signature         varchar(255) not null,
    proposal_block_id bigint not null references blocks(proposal_block_id),
    created_at        timestamptz not null default now()
);

-- +migrate Down

DROP TABLE signatures;
