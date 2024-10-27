package pgx

import (
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"intmax2-node/internal/hash/goldenposeidon"

	"github.com/google/uuid"
)

func (p *pgx) CreateBalanceProof(
	userAddress string,
	blockNumber uint32,
	privateStateCommitment goldenposeidon.PoseidonHashOut,
	balanceProof []byte,
) (*mDBApp.BalanceProof, error) {
	const query = ` INSERT INTO balance_proofs
	(id, user_address, block_number, private_state_commitment, balance_proof, created_at)
	VALUES ($1, $2, $3, $4, $5, $6) `

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(
		query,
		id,
		userAddress,
		blockNumber,
		privateStateCommitment,
		balanceProof,
		createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create balance proof: %w", err)
	}

	return p.GetBalanceProof(id)
}

func (p *pgx) GetBalanceProof(id string) (*mDBApp.BalanceProof, error) {
	const query = ` SELECT id, user_address, block_number, private_state_commitment, balance_proof, created_at
	FROM balance_proofs WHERE id = $1 `

	var b models.BalanceProof
	err := errPgx.Err(p.queryRow(p.ctx, query, id).
		Scan(
			&b.ID,
			&b.UserAddress,
			&b.BlockNumber,
			&b.PrivateStateCommitment,
			&b.BalanceProof,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, fmt.Errorf("failed to get balance proof: %w", err)
	}
	balanceProof := p.balanceProofToDBApp(&b)
	return &balanceProof, nil
}

func (p *pgx) balanceProofToDBApp(
	b *models.BalanceProof,
) mDBApp.BalanceProof {
	return mDBApp.BalanceProof{
		ID:                     b.ID,
		UserAddress:            b.UserAddress,
		BlockNumber:            uint32(b.BlockNumber),
		PrivateStateCommitment: b.PrivateStateCommitment,
		BalanceProof:           b.BalanceProof,
		CreatedAt:              b.CreatedAt,
	}
}
