-- +migrate Up

-- DROP TABLE block_contained_senders;

CREATE TABLE block_participants (
    id           uuid not null default uuid_generate_v4(),
    block_number bigint not null,
    sender_id    uuid not null,
    created_at   timestamptz not null default now(),
    PRIMARY KEY (id),
    FOREIGN KEY (sender_id) REFERENCES senders(id) ON DELETE CASCADE,
    UNIQUE (block_number, sender_id)
);

-- +migrate Down

DROP TABLE block_participants;
-- DROP TABLE block_contained_senders;

-- CREATE TABLE block_contained_senders (
--     id          uuid not null default uuid_generate_v4(),
--     block_hash  varchar(64) not null,
--     sender      varchar(64) not null,
--     PRIMARY KEY (id)
-- );
