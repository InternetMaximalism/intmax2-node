-- +migrate Up

CREATE TABLE deposits (
    id                    uuid not null default uuid_generate_v4(),
    deposit_id            bigint not null,
    deposit_hash          varchar(255) not null,
    recipient_salt_hash   varchar(255) not null,
    token_index           bigint not null,
    amount                varchar(255) not null,
    created_at            timestamptz not null default now(),
    PRIMARY KEY (id)
);

-- +migrate Down

DROP SEQUENCE deposits;
