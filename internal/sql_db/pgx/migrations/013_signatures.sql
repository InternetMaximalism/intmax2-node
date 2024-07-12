-- +migrate Up

ALTER TABLE signatures ADD COLUMN proposal_block_id uuid references blocks(proposal_block_id);

-- +migrate Down

ALTER TABLE signatures DROP COLUMN proposal_block_id;
