-- +migrate Up

ALTER TABLE backup_transactions ADD COLUMN tx_double_hash text;

ALTER TABLE backup_transfers ADD COLUMN transfer_double_hash text;

ALTER TABLE backup_deposits ADD COLUMN deposit_double_hash text;

-- +migrate Down

ALTER TABLE backup_transactions DROP COLUMN tx_double_hash;

ALTER TABLE backup_transfers DROP COLUMN transfer_double_hash;

ALTER TABLE backup_deposits DROP COLUMN deposit_double_hash;
