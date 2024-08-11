-- +migrate Up

ALTER SEQUENCE accounts_account_id_seq RESTART WITH 2;

DELETE FROM event_block_numbers WHERE event_name = 'BlockPosted';

-- +migrate Down

ALTER SEQUENCE accounts_account_id_seq RESTART;

DELETE FROM event_block_numbers WHERE event_name = 'BlockPosted';
