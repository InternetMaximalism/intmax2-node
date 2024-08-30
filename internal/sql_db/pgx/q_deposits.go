package pgx

import (
	"encoding/hex"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	intMaxTree "intmax2-node/internal/tree"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

func (p *pgx) CreateDeposit(
	depositLeaf intMaxTree.DepositLeaf,
	depositId uint32,
) (*mDBApp.Deposit, error) {
	depositHash := depositLeaf.Hash().Hex()
	recipientSaltHash := hex.EncodeToString(depositLeaf.RecipientSaltHash[:])
	tokenIndex := int64(depositLeaf.TokenIndex)
	amount := depositLeaf.Amount.String()

	s := models.Deposit{
		ID:                uuid.New().String(),
		DepositID:         int64(depositId),
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

func (b *pgx) UpdateDepositIndexByDepositHash(depositHash common.Hash, tokenIndex uint32) error {
	const (
		q = `UPDATE deposits SET token_index = $1 WHERE deposit_hash = $2`
	)

	_, err := b.exec(b.ctx, q, tokenIndex, depositHash.Hex())
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) Deposit(ID string) (*mDBApp.Deposit, error) {
	const (
		q = `SELECT
             id ,deposit_index ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at ,deposit_id
             FROM deposits WHERE id = $1`
	)

	var tmp models.Deposit
	err := errPgx.Err(p.queryRow(p.ctx, q, ID).
		Scan(
			&tmp.ID,
			&tmp.DepositIndex,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.CreatedAt,
			&tmp.DepositID,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) ScanDeposits() ([]*mDBApp.Deposit, error) {
	const (
		q = `SELECT
			 id ,deposit_index ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at ,deposit_id
			 FROM deposits`
	)

	rows, err := p.query(p.ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bDBApp []*mDBApp.Deposit
	for rows.Next() {
		var tmp models.Deposit
		err = errPgx.Err(rows.Scan(
			&tmp.ID,
			&tmp.DepositIndex,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.CreatedAt,
			&tmp.DepositID,
		))
		if err != nil {
			return nil, err
		}

		bDBApp = append(bDBApp, p.depositToDBApp(&tmp))
	}

	return bDBApp, nil
}

func (p *pgx) DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error) {
	const (
		q = `SELECT
             id ,deposit_id ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at ,deposit_index
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
			&tmp.DepositIndex,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error) {
	const (
		q = `SELECT
             id ,deposit_id ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at ,deposit_index
             FROM deposits WHERE deposit_hash = $1`
	)

	var tmp models.Deposit
	err := errPgx.Err(p.queryRow(p.ctx, q, depositHash.Hex()).
		Scan(
			&tmp.ID,
			&tmp.DepositID,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.CreatedAt,
			&tmp.DepositIndex,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) depositToDBApp(tmp *models.Deposit) *mDBApp.Deposit {
	depositIndex := new(uint32)
	if tmp.DepositIndex != nil {
		*depositIndex = uint32(*tmp.DepositIndex)
	}
	m := mDBApp.Deposit{
		ID:                tmp.ID,
		DepositID:         uint32(tmp.DepositID),
		DepositIndex:      depositIndex,
		DepositHash:       tmp.DepositHash,
		RecipientSaltHash: tmp.RecipientSaltHash,
		TokenIndex:        uint32(tmp.TokenIndex),
		Amount:            tmp.Amount,
		CreatedAt:         tmp.CreatedAt,
	}

	return &m
}
