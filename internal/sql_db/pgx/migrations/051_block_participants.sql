-- +migrate Up

ALTER TABLE block_participants
    ADD CONSTRAINT block_participants_block_number_fkey
    FOREIGN KEY (block_number) REFERENCES block_contents (block_number) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_block_participants_block_number ON block_participants(block_number);
CREATE INDEX IF NOT EXISTS idx_block_participants_sender_id ON block_participants(sender_id);

-- +migrate Down

ALTER TABLE block_participants
    DROP CONSTRAINT block_participants_block_number_fkey;

DROP INDEX IF EXISTS idx_block_participants_block_number;
DROP INDEX IF EXISTS idx_block_participants_sender_id;