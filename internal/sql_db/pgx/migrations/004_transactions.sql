-- +migrate Up

DROP TYPE IF EXISTS transactions_status;
CREATE TYPE transactions_status AS ENUM ('pending', 'committed', 'confirmed', 'failed');

CREATE TABLE transactions (
    tx_hash           varchar(255) primary key,
    sender_public_key varchar(255) not null,
    signature_id      bigint not null references signatures(signature_id),
    status            transactions_status not null default 'pending'::transactions_status,
    created_at        timestamptz not null default now()
);

-- +migrate Down

DROP TABLE transactions;
DROP TYPE IF EXISTS transactions_status;
