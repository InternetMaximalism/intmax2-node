-- +migrate Up

CREATE TABLE block_proofs (
    id               uuid not null default uuid_generate_v4(),
    block_content_id uuid not null,
    block_proof      bytea not null,
    PRIMARY KEY (id),
    FOREIGN KEY (block_content_id) REFERENCES block_contents(id) ON DELETE CASCADE,
    UNIQUE (block_content_id)
);

-- +migrate Down

DROP TABLE block_proofs;
