-- +migrate Up

CREATE TABLE user_states (
    id uuid not null default uuid_generate_v4(),
    user_address varchar(66) not null,
    encrypted_user_state text not null,
    auth_signature text not null,
    created_at timestamptz not null default now(),
    modified_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

CREATE TABLE balance_proofs (
    id uuid not null default uuid_generate_v4(),
    user_address varchar(66) not null,
    private_state_commitment varchar(66) not null,
    block_number int not null,
    balance_proof bytea not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id),
    UNIQUE (user_address, block_number, private_state_commitment)
);

CREATE TABLE spent_proofs (
    id uuid not null default uuid_generate_v4(),
    spent_proof bytea not null,
    created_at timestamptz not null default now(),
    PRIMARY KEY (id)
);

-- +migrate Down

DROP TABLE spent_proofs;
DROP TABLE balance_proofs;
DROP TABLE user_states;
