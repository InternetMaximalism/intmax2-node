-- +migrate Up

ALTER TABLE deposits
ADD CONSTRAINT unique_deposits_deposit_index UNIQUE (deposit_index);

-- +migrate Down

ALTER TABLE deposits
DROP CONSTRAINT unique_deposits_deposit_index;
