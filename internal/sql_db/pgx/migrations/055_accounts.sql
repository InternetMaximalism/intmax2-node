-- +migrate Up

DELETE FROM event_block_numbers WHERE event_name = 'BlockPosted';
DELETE FROM accounts WHERE 1=1;
DELETE FROM senders WHERE 1=1;
ALTER SEQUENCE accounts_account_id_seq RESTART WITH 2;

-- +migrate Down

DELETE FROM event_block_numbers WHERE event_name = 'BlockPosted';
DELETE FROM accounts WHERE 1=1;
DELETE FROM senders WHERE 1=1;
ALTER SEQUENCE accounts_account_id_seq RESTART WITH 2;
