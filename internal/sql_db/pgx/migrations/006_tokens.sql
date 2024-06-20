-- +migrate Up

CREATE TABLE tokens (
    token_index   varchar(255) not null primary key,
    token_address varchar(255),
    token_id      bigint not null,
    created_at    timestamptz not null default now()
);

-- +migrate Down

DROP TABLE tokens;
