-- +migrate Up

ALTER TABLE deposits ADD COLUMN sender uuid references ethereum_counterparties(id);

CREATE INDEX idx_deposits_sender ON deposits(sender);

-- +migrate Down

ALTER TABLE deposits DROP COLUMN sender;
