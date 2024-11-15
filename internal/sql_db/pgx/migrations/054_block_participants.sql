-- +migrate Up

ALTER TABLE block_participants
    DROP CONSTRAINT block_participants_sender_id_fkey;

DELETE FROM relationship_l2_batch_index_block_contents WHERE 1=1;
DELETE FROM l2_batch_index WHERE 1=1;
DELETE FROM deposits WHERE 1=1;
DELETE FROM block_contents WHERE 1=1;
DELETE FROM event_block_numbers_validity_prover WHERE 1=1;
DELETE FROM block_accounts WHERE 1=1;
DELETE FROM block_senders WHERE 1=1;
ALTER SEQUENCE block_accounts_account_id_seq RESTART WITH 2;

ALTER TABLE block_participants
    ADD CONSTRAINT block_participants_sender_id_fkey
        FOREIGN KEY (sender_id) REFERENCES block_senders(id) ON DELETE CASCADE;

-- +migrate Down

ALTER TABLE block_participants
    DROP CONSTRAINT block_participants_sender_id_fkey;

DELETE FROM relationship_l2_batch_index_block_contents WHERE 1=1;
DELETE FROM l2_batch_index WHERE 1=1;
DELETE FROM deposits WHERE 1=1;
DELETE FROM block_contents WHERE 1=1;
DELETE FROM event_block_numbers_validity_prover WHERE 1=1;
DELETE FROM block_accounts WHERE 1=1;
DELETE FROM block_senders WHERE 1=1;
ALTER SEQUENCE block_accounts_account_id_seq RESTART WITH 2;

ALTER TABLE block_participants
    ADD CONSTRAINT block_participants_sender_id_fkey
        FOREIGN KEY (sender_id) REFERENCES senders(id) ON DELETE CASCADE;
