-- +migrate Up

DROP TABLE backup_balances;

CREATE TABLE backup_balances (
    id uuid not null default uuid_generate_v4(),
    user_address VARCHAR(66) NOT NULL,
    encrypted_balance_proof TEXT NOT NULL, -- not encrypted
    encrypted_balance_data TEXT NOT NULL,
    encrypted_txs json NOT NULL, -- sent_txs
    encrypted_transfers json NOT NULL, -- received_transfers
    encrypted_deposits json NOT NULL,
    block_number int not null,
    signature TEXT NOT NULL, -- not need to save
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE INDEX idx_backup_balances_user_address ON backup_balances(user_address);

-- +migrate Down

DELETE FROM backup_balances WHERE LENGTH(user_address) > 42;
ALTER TABLE backup_balances ALTER COLUMN user_address TYPE VARCHAR(42);
