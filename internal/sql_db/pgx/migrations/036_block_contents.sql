-- +migrate Up

CREATE TABLE block_validity_proofs (
    id               uuid not null default uuid_generate_v4(),
    block_content_id uuid not null,
    validity_proof      bytea not null,
    PRIMARY KEY (id),
    FOREIGN KEY (block_content_id) REFERENCES block_contents(id) ON DELETE CASCADE,
    UNIQUE (block_content_id)
);

-- +migrate Down

DROP TABLE block_validity_proofs;
