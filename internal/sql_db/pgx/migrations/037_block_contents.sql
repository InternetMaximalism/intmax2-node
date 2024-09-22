-- +migrate Up

ALTER TABLE block_contents
ADD CONSTRAINT unique_block_contents_block_number UNIQUE (block_number);

-- +migrate Down

ALTER TABLE block_contents
DROP CONSTRAINT unique_block_contents_block_number;
