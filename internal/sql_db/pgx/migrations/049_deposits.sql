-- +migrate Up

ALTER TABLE deposits RENAME COLUMN amount TO amount_old;
ALTER TABLE deposits ADD COLUMN amount numeric not null default 0;
UPDATE deposits SET amount = amount_old::numeric WHERE 1=1;
ALTER TABLE deposits DROP COLUMN amount_old;

-- +migrate Down

ALTER TABLE deposits RENAME COLUMN amount TO amount_old;
ALTER TABLE deposits ADD COLUMN amount varchar(255) not null default '0';
UPDATE deposits SET amount = amount_old WHERE 1=1;
ALTER TABLE deposits DROP COLUMN amount_old;
