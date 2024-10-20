-- +migrate Up

CREATE TABLE relationship_l2_batch_index_block_contents (
  l2_batch_index numeric not null references l2_batch_index(l2_batch_index),
  block_contents_id uuid not null references block_contents(id),
  created_at timestamptz not null default now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_relationship_l2_batch_index_block_contents
    ON relationship_l2_batch_index_block_contents(l2_batch_index, block_contents_id);

-- +migrate Down

DROP TABLE relationship_l2_batch_index_block_contents;
