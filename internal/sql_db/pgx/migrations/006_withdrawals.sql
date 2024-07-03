-- +migrate Up

CREATE TABLE withdrawals (
    id uuid not null default uuid_generate_v4(),
    status int not null,
    recipient varchar(64) not null,
    token_index int not null,
    amount varchar(32) not null,
    salt varchar(32) not null,
    transfer_hash varchar(32) not null,
    transfer_merkle_proof json not null,
    transaction json not null,
    tx_merkle_proof json not null,
    block_number int not null,
    enough_balance_proof varchar(255) not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

-- +migrate Down

DROP TABLE withdrawals;
