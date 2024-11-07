package pgx

import (
	"encoding/hex"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	intMaxTree "intmax2-node/internal/tree"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

func (p *pgx) CreateDeposit(
	depositLeaf intMaxTree.DepositLeaf,
	depositId uint32,
	sender string,
) (*mDBApp.Deposit, error) {
	depositHash := depositLeaf.Hash().Hex()
	recipientSaltHash := hex.EncodeToString(depositLeaf.RecipientSaltHash[:])
	tokenIndex := int64(depositLeaf.TokenIndex)

	var amount uint256.Int
	_ = amount.SetFromBig(depositLeaf.Amount)
	amountV, _ := amount.Value()

	const (
		q = `INSERT INTO deposits (
             deposit_id ,deposit_hash ,recipient_salt_hash ,token_index ,amount ,sender
             ) VALUES ($1, $2, $3, $4, $5, $6) `
	)

	_, err := p.exec(p.ctx, q,
		depositId,
		depositHash,
		recipientSaltHash,
		tokenIndex,
		amountV,
		sender,
	)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.Deposit
	bDBApp, err = p.DepositByDepositID(depositId)
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

func (p *pgx) UpdateSenderByDepositID(depositID uint32, sender string) error {
	const (
		q = `UPDATE deposits SET sender = $1 WHERE deposit_id = $2`
	)

	_, err := p.exec(p.ctx, q, sender, depositID)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) Deposit(Id string) (*mDBApp.Deposit, error) {
	const (
		q = `
SELECT
d.id
,d.deposit_id
,d.deposit_hash
,d.recipient_salt_hash
,d.token_index
,d.amount
,d.deposit_index
,ec.address AS sender
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter > d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_after_deposit_index
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter <= d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_before_deposit_index
,(CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT bc.block_number
    FROM block_contents bc
    LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
    WHERE bp.validity_proof IS NOT NULL AND bc.deposit_leaves_counter > d.deposit_index
    ORDER BY bc.block_number ASC
    LIMIT 1)
    , 0)
)
ELSE 0
END) is_sync
,d.created_at
FROM deposits d
LEFT JOIN ethereum_counterparties ec on ec.id = d.sender
WHERE d.id = $1
`
	)

	var tmp models.Deposit
	err := errPgx.Err(p.queryRow(p.ctx, q, Id).
		Scan(
			&tmp.ID,
			&tmp.DepositID,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.DepositIndex,
			&tmp.Sender,
			&tmp.BlockNumberAfterDepositIndex,
			&tmp.BlockNumberBeforeDepositIndex,
			&tmp.IsSync,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) ScanDeposits() ([]*mDBApp.Deposit, error) {
	const (
		q = `
SELECT
d.id
,d.deposit_id
,d.deposit_hash
,d.recipient_salt_hash
,d.token_index
,d.amount
,d.deposit_index
,ec.address AS sender
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter > d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_after_deposit_index
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter <= d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_before_deposit_index
,(CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT bc.block_number
    FROM block_contents bc
    LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
    WHERE bp.validity_proof IS NOT NULL AND bc.deposit_leaves_counter > d.deposit_index
    ORDER BY bc.block_number ASC
    LIMIT 1)
    , 0)
)
ELSE 0
END) is_sync
,d.created_at
FROM deposits d
LEFT JOIN ethereum_counterparties ec on ec.id = d.sender
WHERE d.deposit_index IS NOT NULL ORDER BY d.deposit_index ASC
`
	)

	rows, err := p.query(p.ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var bDBApp []*mDBApp.Deposit
	for rows.Next() {
		var tmp models.Deposit
		err = errPgx.Err(rows.Scan(
			&tmp.ID,
			&tmp.DepositID,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.DepositIndex,
			&tmp.Sender,
			&tmp.BlockNumberAfterDepositIndex,
			&tmp.BlockNumberBeforeDepositIndex,
			&tmp.IsSync,
			&tmp.CreatedAt,
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
		q = `
SELECT
d.id
,d.deposit_id
,d.deposit_hash
,d.recipient_salt_hash
,d.token_index
,d.amount
,d.deposit_index
,ec.address AS sender
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter > d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_after_deposit_index
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter <= d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_before_deposit_index
,(CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT bc.block_number
    FROM block_contents bc
    LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
    WHERE bp.validity_proof IS NOT NULL AND bc.deposit_leaves_counter > d.deposit_index
    ORDER BY bc.block_number ASC
    LIMIT 1)
    , 0)
)
ELSE 0
END) is_sync
,d.created_at
FROM deposits d
LEFT JOIN ethereum_counterparties ec on ec.id = d.sender
WHERE d.deposit_id = $1
`
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
			&tmp.DepositIndex,
			&tmp.Sender,
			&tmp.BlockNumberAfterDepositIndex,
			&tmp.BlockNumberBeforeDepositIndex,
			&tmp.IsSync,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error) {
	const (
		q = `
SELECT
d.id
,d.deposit_id
,d.deposit_hash
,d.recipient_salt_hash
,d.token_index
,d.amount
,d.deposit_index
,ec.address AS sender
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter > d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_after_deposit_index
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter <= d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_before_deposit_index
,(CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT bc.block_number
    FROM block_contents bc
    LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
    WHERE bp.validity_proof IS NOT NULL AND bc.deposit_leaves_counter > d.deposit_index
    ORDER BY bc.block_number ASC
    LIMIT 1)
    , 0)
)
ELSE 0
END) is_sync
,d.created_at
FROM deposits d
LEFT JOIN ethereum_counterparties ec on ec.id = d.sender
WHERE d.deposit_hash = $1
`
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
			&tmp.DepositIndex,
			&tmp.Sender,
			&tmp.BlockNumberAfterDepositIndex,
			&tmp.BlockNumberBeforeDepositIndex,
			&tmp.IsSync,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.depositToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) DepositsListByDepositHash(depositHash ...common.Hash) ([]*mDBApp.Deposit, error) {
	const (
		q = `
SELECT
d.id
,d.deposit_id
,d.deposit_hash
,d.recipient_salt_hash
,d.token_index
,d.amount
,d.deposit_index
,ec.address AS sender
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter > d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_after_deposit_index
, (CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT block_number FROM block_contents where deposit_leaves_counter <= d.deposit_index
    ORDER by block_number ASC limit 1)
    ,0)
)
ELSE 0
END) block_number_before_deposit_index
,(CASE
WHEN d.deposit_index IS NOT NULL THEN (
    COALESCE(
    (SELECT bc.block_number
    FROM block_contents bc
    LEFT JOIN block_validity_proofs bp ON bc.id = bp.block_content_id
    WHERE bp.validity_proof IS NOT NULL AND bc.deposit_leaves_counter > d.deposit_index
    ORDER BY bc.block_number ASC
    LIMIT 1)
    , 0)
)
ELSE 0
END) is_sync
,d.created_at
FROM deposits d
LEFT JOIN ethereum_counterparties ec on ec.id = d.sender
WHERE d.deposit_hash = any($1)
ORDER BY d.deposit_id ASC `
	)

	hashes := make([]string, len(depositHash))
	for key := range depositHash {
		hashes[key] = depositHash[key].String()
	}

	rows, err := p.query(p.ctx, q, hashes)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var bDBApp []*mDBApp.Deposit
	for rows.Next() {
		var tmp models.Deposit
		err = errPgx.Err(rows.Scan(
			&tmp.ID,
			&tmp.DepositID,
			&tmp.DepositHash,
			&tmp.RecipientSaltHash,
			&tmp.TokenIndex,
			&tmp.Amount,
			&tmp.DepositIndex,
			&tmp.Sender,
			&tmp.BlockNumberAfterDepositIndex,
			&tmp.BlockNumberBeforeDepositIndex,
			&tmp.IsSync,
			&tmp.CreatedAt,
		))
		if err != nil {
			return nil, err
		}

		bDBApp = append(bDBApp, p.depositToDBApp(&tmp))
	}

	return bDBApp, nil
}

func (p *pgx) FetchNextDepositIndex() (uint32, error) {
	const (
		q = `SELECT COALESCE(MAX(deposit_index) + 1, 0) FROM deposits`
	)

	var nextDepositIndex uint32
	err := p.queryRow(p.ctx, q).Scan(&nextDepositIndex)
	if err != nil {
		return 0, fmt.Errorf("FetchNextDepositIndex error: %w", err)
	}

	return nextDepositIndex, nil
}

const int32Key = 32

func (p *pgx) depositToDBApp(tmp *models.Deposit) *mDBApp.Deposit {
	depositIndex := new(uint32)
	if tmp.DepositIndex != nil {
		*depositIndex = uint32(*tmp.DepositIndex)
	}

	m := mDBApp.Deposit{
		ID:                            tmp.ID,
		DepositID:                     uint32(tmp.DepositID),
		DepositIndex:                  depositIndex,
		DepositHash:                   common.HexToHash(tmp.DepositHash),
		RecipientSaltHash:             [int32Key]byte(common.HexToHash("0x" + tmp.RecipientSaltHash)),
		TokenIndex:                    uint32(tmp.TokenIndex),
		Amount:                        tmp.Amount.ToBig(),
		Sender:                        tmp.Sender.String,
		BlockNumberAfterDepositIndex:  tmp.BlockNumberAfterDepositIndex,
		BlockNumberBeforeDepositIndex: tmp.BlockNumberBeforeDepositIndex,
		CreatedAt:                     tmp.CreatedAt,
	}

	if tmp.IsSync > 0 {
		m.IsSync = true
	}

	return &m
}
