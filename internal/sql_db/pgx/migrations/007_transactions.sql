-- +migrate Up

ALTER TABLE transactions ALTER COLUMN tx_hash SET NOT NULL;

-- +migrate Down

ALTER TABLE transactions ALTER COLUMN tx_hash DROP NOT NULL;
