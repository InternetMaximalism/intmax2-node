-- +migrate Up

ALTER TABLE backup_balances ALTER COLUMN user_address TYPE VARCHAR(66) USING user_address::VARCHAR(66);

-- +migrate Down

DELETE FROM backup_balances WHERE LENGTH(user_address) > 42;
ALTER TABLE backup_balances ALTER COLUMN user_address TYPE VARCHAR(42);
