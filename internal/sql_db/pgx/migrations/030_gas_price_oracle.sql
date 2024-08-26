-- +migrate Up

CREATE TABLE gas_price_oracle (
    gas_price_oracle_name varchar(255) not null unique,
    value numeric not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (gas_price_oracle_name),
    CONSTRAINT check_gas_price_oracle_v_more_than_zero CHECK (value > 0)
);

-- +migrate Down

DROP TABLE gas_price_oracle;
