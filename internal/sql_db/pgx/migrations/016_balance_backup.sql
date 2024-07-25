-- +migrate Up

CREATE TABLE IF NOT EXISTS balance_backup (
    id                          UUID            NOT NULL        DEFAULT     uuid_generate_v4(),
    user_address                VARCHAR(255)    NOT NULL,
    block_number                INT             NOT NULL,
    encrypted_balance_proof     VARCHAR(255)    NOT NULL,
    encrypted_public_inputs     VARCHAR(255)    NOT NULL,
    encrypted_txs               JSONB           NOT NULL,
    encrypted_transfers         JSONB           NOT NULL,
    encrypted_deposits          JSONB           NOT NULL,
    signature                   VARCHAR(255)    NOT NULL,
    created_at                  TIMESTAMPTZ     NOT NULL        DEFAULT     NOW(),

    PRIMARY KEY (id),
    UNIQUE (user_address, block_number)
);

-- +migrate Down

DROP TABLE IF EXISTS balance_backup;
