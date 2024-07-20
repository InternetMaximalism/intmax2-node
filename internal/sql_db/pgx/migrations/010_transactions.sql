-- +migrate Up

ALTER TABLE transactions ADD COLUMN tx_tree_info jsonb NOT NULL default '{}'::jsonb;
ALTER TABLE transactions RENAME status TO status_old;
ALTER TABLE transactions ADD COLUMN status int;
ALTER TABLE transactions DROP COLUMN status_old;
DROP TYPE IF EXISTS transactions_status;

-- +migrate Down

DROP TYPE IF EXISTS transactions_status;
CREATE TYPE transactions_status AS ENUM ('pending', 'committed', 'confirmed', 'failed');

ALTER TABLE transactions RENAME status TO status_old;
ALTER TABLE transactions ADD COLUMN status transactions_status not null default 'pending'::transactions_status;
ALTER TABLE transactions DROP COLUMN status_old;
ALTER TABLE transactions DROP COLUMN tx_tree_info;
