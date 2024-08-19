-- +migrate Up

ALTER TABLE backup_transactions ADD COLUMN tx_double_hash text;
CREATE INDEX idx_backup_transactions_created_at_block_number ON backup_transactions(created_at, block_number);
CREATE INDEX idx_backup_transactions_created_at_id ON backup_transactions(created_at, id);

ALTER TABLE backup_transfers ADD COLUMN transfer_double_hash text;
CREATE INDEX idx_backup_transfers_created_at_block_number ON backup_transfers(created_at, block_number);
CREATE INDEX idx_backup_transfers_created_at_id ON backup_transfers(created_at, id);

ALTER TABLE backup_deposits ADD COLUMN deposit_double_hash text;
CREATE INDEX idx_backup_deposits_created_at_block_number ON backup_deposits(created_at, block_number);
CREATE INDEX idx_backup_deposits_created_at_id ON backup_deposits(created_at, id);

-- +migrate Down

ALTER TABLE backup_transactions DROP COLUMN tx_double_hash;
DROP INDEX idx_backup_transactions_created_at_block_number;
DROP INDEX idx_backup_transactions_created_at_id;

ALTER TABLE backup_transfers DROP COLUMN transfer_double_hash;
DROP INDEX idx_backup_transfers_created_at_block_number;
DROP INDEX idx_backup_transfers_created_at_id;

ALTER TABLE backup_deposits DROP COLUMN deposit_double_hash;
DROP INDEX idx_backup_deposits_created_at_block_number;
DROP INDEX idx_backup_deposits_created_at_id;
