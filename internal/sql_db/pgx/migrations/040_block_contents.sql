-- +migrate Up

ALTER TABLE block_contents ADD COLUMN block_number_l2 numeric;
ALTER TABLE block_contents ADD COLUMN block_hash_l2 varchar;

-- +migrate Down

ALTER TABLE block_contents DROP COLUMN block_hash_l2;
ALTER TABLE block_contents DROP COLUMN block_number_l2;
