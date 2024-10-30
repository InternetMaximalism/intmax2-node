package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/google/uuid"
)

func (p *pgx) CreateBalanceProof(
	userStateID, userAddress, privateStateCommitment string,
	blockNumber int64,
	balanceProof []byte,
) (*mDBApp.BalanceProof, error) {
	const (
		q = ` INSERT INTO balance_proofs (
              id ,user_state_id ,user_address ,block_number ,private_state_commitment ,balance_proof
              ) VALUES ($1, $2, $3, $4, $5, $6) `
	)

	id := uuid.New().String()

	err := p.createBackupEntry(q,
		id,
		userStateID,
		userAddress,
		blockNumber,
		privateStateCommitment,
		balanceProof,
	)
	if err != nil {
		return nil, err
	}

	return p.GetBalanceProof(id)
}

func (p *pgx) GetBalanceProof(id string) (*mDBApp.BalanceProof, error) {
	const (
		q = ` SELECT id
              ,user_state_id ,user_address ,block_number ,private_state_commitment
              ,balance_proof ,created_at ,updated_at
              FROM balance_proofs WHERE id = $1 `
	)

	var b models.BalanceProof
	err := errPgx.Err(p.queryRow(p.ctx, q, id).
		Scan(
			&b.ID,
			&b.UserStateID,
			&b.UserAddress,
			&b.BlockNumber,
			&b.PrivateStateCommitment,
			&b.BalanceProof,
			&b.CreatedAt,
			&b.UpdatedAt,
		))
	if err != nil {
		return nil, err
	}

	balanceProof := p.balanceProofToDBApp(&b)
	return &balanceProof, nil
}

func (p *pgx) GetBalanceProofByUserStateID(userStateID string) (*mDBApp.BalanceProof, error) {
	const (
		q = ` SELECT id
              ,user_state_id ,user_address ,block_number ,private_state_commitment
              ,balance_proof ,created_at ,updated_at
              FROM balance_proofs WHERE user_state_id = $1 `
	)

	var b models.BalanceProof
	err := errPgx.Err(p.queryRow(p.ctx, q, userStateID).
		Scan(
			&b.ID,
			&b.UserStateID,
			&b.UserAddress,
			&b.BlockNumber,
			&b.PrivateStateCommitment,
			&b.BalanceProof,
			&b.CreatedAt,
			&b.UpdatedAt,
		))
	if err != nil {
		return nil, err
	}

	balanceProof := p.balanceProofToDBApp(&b)
	return &balanceProof, nil
}

func (p *pgx) balanceProofToDBApp(b *models.BalanceProof) mDBApp.BalanceProof {
	return mDBApp.BalanceProof{
		ID:                     b.ID,
		UserStateID:            b.UserStateID,
		UserAddress:            b.UserAddress,
		BlockNumber:            b.BlockNumber,
		PrivateStateCommitment: b.PrivateStateCommitment,
		BalanceProof:           b.BalanceProof,
		CreatedAt:              b.CreatedAt,
		UpdatedAt:              b.UpdatedAt,
	}
}
