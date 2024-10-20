package pgx

import (
	"encoding/hex"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	intMaxTree "intmax2-node/internal/tree"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"math/big"
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

func (p *pgx) UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error {
	const (
		q = `UPDATE deposits SET deposit_index = $1 WHERE deposit_hash = $2`
	)

	_, err := p.exec(p.ctx, q, depositIndex, depositHash.Hex())
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) Deposit(Id string) (*mDBApp.Deposit, error) {
	const (
		q = `SELECT
             id ,deposit_index ,deposit_hash ,recipient_salt_hash
			 ,token_index ,amount ,created_at ,deposit_id
             FROM deposits WHERE id = $1`
	)

	var tmp models.Deposit
	err := errPgx.Err(p.queryRow(p.ctx, q, Id).
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
			 FROM deposits WHERE deposit_index IS NOT NULL ORDER BY deposit_index ASC`
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

func (p *pgx) FetchLastDepositIndex() (uint32, error) {
	const (
		q = `SELECT MAX(deposit_index) FROM deposits`
	)

	var lastDepositIndex *uint32
	err := errPgx.Err(p.queryRow(p.ctx, q).Scan(&lastDepositIndex))
	if err != nil {
		// if errors.Is(err, pgxV5.ErrNoRows) {
		// 	return 0, nil
		// }

		fmt.Printf("FetchLastDepositIndex error: %v\n", err)

		return 0, err
	}

	if lastDepositIndex == nil {
		return 0, nil
	}

	return *lastDepositIndex, nil
}

const int32Key = 32

func (p *pgx) depositToDBApp(tmp *models.Deposit) *mDBApp.Deposit {
	depositIndex := new(uint32)
	if tmp.DepositIndex != nil {
		*depositIndex = uint32(*tmp.DepositIndex)
	}

	amount, ok := new(big.Int).SetString(tmp.Amount, 10)
	if !ok {
		// Fatal error
		panic("depositToDBApp: invalid number string")
	}

	m := mDBApp.Deposit{
		ID:                tmp.ID,
		DepositID:         uint32(tmp.DepositID),
		DepositIndex:      depositIndex,
		DepositHash:       common.HexToHash("0x" + tmp.DepositHash),
		RecipientSaltHash: [int32Key]byte(common.HexToHash("0x" + tmp.RecipientSaltHash)),
		TokenIndex:        uint32(tmp.TokenIndex),
		Amount:            amount,
		CreatedAt:         tmp.CreatedAt,
	}

	return &m
}
