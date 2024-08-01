-- +migrate Up

ALTER TABLE blocks ADD COLUMN senders json not null default '[]';

-- +migrate Down

ALTER TABLE blocks DROP COLUMN senders;
