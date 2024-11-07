-- +migrate Up

ALTER TABLE block_contents ADD COLUMN deposit_leaves_counter bigint not null default 0;

CREATE INDEX idx_block_contents_deposit_leaves_counter ON block_contents(deposit_leaves_counter);

-- +migrate Down

ALTER TABLE block_contents DROP COLUMN deposit_leaves_counter;
