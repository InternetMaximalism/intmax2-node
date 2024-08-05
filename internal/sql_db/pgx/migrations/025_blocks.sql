-- +migrate Up

ALTER TABLE blocks ADD COLUMN block_number int;

-- +migrate Down

ALTER TABLE blocks DROP COLUMN block_number;
