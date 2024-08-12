-- +migrate Up

CREATE TABLE backup_balances (
    id uuid not null default uuid_generate_v4(),
    user_address VARCHAR(42) NOT NULL,
    encrypted_balance_proof TEXT NOT NULL,
    encrypted_balance_data TEXT NOT NULL,
    encrypted_txs json NOT NULL,
    encrypted_transfers json NOT NULL,
    encrypted_deposits json NOT NULL,
    signature TEXT NOT NULL,
    created_at timestamptz not null default now()
);

CREATE INDEX idx_backup_balances_user_address ON backup_balances(user_address);

-- +migrate Down

DROP TABLE backup_balances;