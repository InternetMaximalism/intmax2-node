-- +migrate Up

CREATE TABLE event_block_numbers (
    id UUID not null default uuid_generate_v4(),
    event_name varchar(255) not null unique,
    last_processed_block_number bigint not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id),
    CONSTRAINT check_last_processed_block_number_positive CHECK (last_processed_block_number >= 0)
);

CREATE INDEX idx_event_block_numbers_event_name ON event_block_numbers(event_name);

-- +migrate Down

DROP TABLE event_block_numbers;
