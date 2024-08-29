package pgx

import (
	"encoding/hex"
	"intmax2-node/internal/block_validity_prover"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateDeposit(
	deposit block_validity_prover.DepositLeafWithId,
) (*mDBApp.Deposit, error) {
	depositID := int64(deposit.DepositId)
	depositHash := deposit.DepositLeaf.Hash().Hex()
	recipientSaltHash := hex.EncodeToString(deposit.DepositLeaf.RecipientSaltHash[:])
	tokenIndex := int64(deposit.DepositLeaf.TokenIndex)
	amount := deposit.DepositLeaf.Amount.String()

	s := models.Deposit{
		ID:                uuid.New().String(),
		DepositID:         depositID,
		DepositHash:       depositHash,
		RecipientSaltHash: recipientSaltHash,
		TokenIndex:        tokenIndex,
		Amount:            amount,
		CreatedAt:         time.Now().UTC(),
	}

	const (
		q = `INSERT INTO deposits (
             id ,deposit_id ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at
             ) VALUES ($1, $2, $3, $4, $5, $6, $7) `
	)

	_, err := p.exec(p.ctx, q,
		s.ID, s.DepositID, s.DepositHash, s.RecipientSaltHash,
		s.TokenIndex, s.Amount, s.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.Deposit
	bDBApp, err = p.Deposit(s.ID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) Deposit(ID string) (*mDBApp.Deposit, error) {
	const (
		q = `SELECT
             id ,deposit_id ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at
             FROM deposits WHERE id = $1`
	)

	var tmp models.Deposit
	err := errPgx.Err(p.queryRow(p.ctx, q, ID).
		Scan(
			&tmp.ID,
			&tmp.DepositID,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error) {
	const (
		q = `SELECT
             id ,deposit_id ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at
             FROM deposits WHERE deposit_id = $1`
	)

	var tmp models.Deposit
	err := errPgx.Err(p.queryRow(p.ctx, q, depositID).
		Scan(
			&tmp.ID,
			&tmp.DepositID,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) depositToDBApp(tmp *models.Deposit) *mDBApp.Deposit {
	m := mDBApp.Deposit{
		ID:                tmp.ID,
		DepositID:         uint32(tmp.DepositID),
		DepositHash:       tmp.DepositHash,
		RecipientSaltHash: tmp.RecipientSaltHash,
		TokenIndex:        uint32(tmp.TokenIndex),
		Amount:            tmp.Amount,
		CreatedAt:         tmp.CreatedAt,
	}

	return &m
}
