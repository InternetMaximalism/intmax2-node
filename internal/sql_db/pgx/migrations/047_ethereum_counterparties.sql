-- +migrate Up

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE ethereum_counterparties (
  id         uuid not null default uuid_generate_v4(),
  address    varchar(255) not null,
  created_at timestamptz not null default now(),
  PRIMARY KEY (id),
  UNIQUE (address)
);

-- +migrate Down

DROP TABLE ethereum_counterparties;
