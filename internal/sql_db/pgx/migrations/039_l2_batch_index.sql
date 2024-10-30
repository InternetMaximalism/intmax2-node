-- +migrate Up

CREATE TABLE l2_batch_index (
  l2_batch_index numeric not null,
  options jsonb not null default '{}'::jsonb,
  l1_verified_batch_tx_hash varchar,
  created_at timestamptz not null default now(),
  PRIMARY KEY (l2_batch_index)
);

-- +migrate Down

DROP TABLE l2_batch_index;
