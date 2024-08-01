-- +migrate Up

CREATE TABLE backup_transfers (
    id uuid not null default uuid_generate_v4(),
    recipient varchar(255) not null,
    encrypted_transfer text not null,
    block_number int not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE TABLE backup_transactions (
    id uuid not null default uuid_generate_v4(),
    sender varchar(255) not null,
    encrypted_tx text not null,
    block_number int not null,
    signature text not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE TABLE backup_deposits (
    id uuid not null default uuid_generate_v4(),
    recipient varchar(255) not null,
    encrypted_deposit text not null,
    block_number int not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE INDEX idx_backup_transfers_recipient ON backup_transfers (recipient);
CREATE INDEX idx_backup_transactions_sender ON backup_transactions (sender);
CREATE INDEX idx_backup_deposits_recipient ON backup_deposits (recipient);

-- +migrate Down

DROP TABLE backup_transfers;
DROP TABLE backup_transactions;
DROP TABLE backup_deposits;