-- +migrate Up

CREATE TABLE ctrl_event_block_numbers_jobs (
    event_name varchar(255) not null unique,
    created_at timestamptz not null default now(),
    PRIMARY KEY (event_name)
);

-- +migrate Down

DROP TABLE ctrl_event_block_numbers_jobs;
