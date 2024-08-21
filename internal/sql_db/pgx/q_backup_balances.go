package pgx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupBalance(
	user, encryptedBalanceProof, encryptedBalanceData, signature string,
	encryptedTxs, encryptedTransfers, encryptedDeposits []string,
	blockNumber int64,
) (*mDBApp.BackupBalance, error) {
	const (
		emptyJSONKey = `[]`

		query = ` INSERT INTO backup_balances (
                  id ,user_address ,encrypted_balance_proof ,encrypted_balance_data
                  ,encrypted_txs ,encrypted_transfers ,encrypted_deposits ,signature
                  ,block_number ,created_at
                  ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) `
	)
	id := uuid.New().String()
	createdAt := time.Now().UTC()

	var (
		err                    error
		encryptedTxsJSON       json.RawMessage
		encryptedTransfersJSON json.RawMessage
		encryptedDepositsJSON  json.RawMessage
	)
	if encryptedTxs == nil {
		encryptedTxsJSON = json.RawMessage(emptyJSONKey)
	} else {
		encryptedTxsJSON, err = json.Marshal(encryptedTxs)
		if err != nil {
			const msg = "error encoding EncryptedTxs: %w"
			return nil, fmt.Errorf(msg, err)
		}
	}

	if encryptedTransfers == nil {
		encryptedTransfersJSON = json.RawMessage(emptyJSONKey)
	} else {
		encryptedTransfersJSON, err = json.Marshal(encryptedTransfers)
		if err != nil {
			const msg = "error encoding EncryptedTransfers: %w"
			return nil, fmt.Errorf(msg, err)
		}
	}

	if encryptedDeposits == nil {
		encryptedDepositsJSON = json.RawMessage(emptyJSONKey)
	} else {
		encryptedDepositsJSON, err = json.Marshal(encryptedDeposits)
		if err != nil {
			const msg = "error encoding EncryptedDeposits: %w"
			return nil, fmt.Errorf(msg, err)
		}
	}

	_, err = p.db.Exec(query,
		id,
		user,
		encryptedBalanceProof,
		encryptedBalanceData,
		encryptedTxsJSON,
		encryptedTransfersJSON,
		encryptedDepositsJSON,
		signature,
		blockNumber,
		createdAt,
	)
	if err != nil {
		const msg = "failed to create backup balance: %w"
		return nil, fmt.Errorf(msg, errPgx.Err(err))
	}

	return p.GetBackupBalance([]string{"id"}, []interface{}{id})
}

func (p *pgx) GetBackupBalance(conditions []string, values []interface{}) (*mDBApp.BackupBalance, error) {
	const (
		baseQuery = `
SELECT
id ,user_address ,encrypted_balance_proof ,encrypted_balance_data
,encrypted_txs ,encrypted_transfers ,encrypted_deposits
,signature ,block_number ,created_at
FROM backup_balances 
WHERE %s
`
	)

	whereClause := make([]string, len(conditions))
	for i, condition := range conditions {
		whereClause[i] = fmt.Sprintf("%s = $%d", condition, i+1)
	}
	query := fmt.Sprintf(baseQuery, strings.Join(whereClause, " AND "))

	var (
		b models.BackupBalance

		encryptedTxs, encryptedTransfers, encryptedDeposits []byte
	)
	err := errPgx.Err(p.queryRow(p.ctx, query, values...).
		Scan(
			&b.ID,
			&b.UserAddress,
			&b.EncryptedBalanceProof,
			&b.EncryptedBalanceData,
			&encryptedTxs,
			&encryptedTransfers,
			&encryptedDeposits,
			&b.Signature,
			&b.BlockNumber,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	err = unmarshalBackupBalanceData(&b, encryptedTxs, encryptedTransfers, encryptedDeposits)
	if err != nil {
		return nil, err
	}

	balance := p.backupBalanceToDBApp(&b)

	return &balance, nil
}

func (p *pgx) GetBackupBalances(condition string, value interface{}) ([]*mDBApp.BackupBalance, error) {
	const (
		baseQuery = `
        SELECT
		    id, user_address, encrypted_balance_proof, encrypted_balance_data,
			encrypted_txs, encrypted_transfers, encrypted_deposits,
			signature, block_number, created_at
        FROM backup_balances
        WHERE %s = $1
`
	)

	query := fmt.Sprintf(baseQuery, condition)

	var balances []*mDBApp.BackupBalance
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) (err error) {
		var (
			b models.BackupBalance

			encryptedTxs, encryptedTransfers, encryptedDeposits []byte
		)
		err = rows.Scan(
			&b.ID,
			&b.UserAddress,
			&b.EncryptedBalanceProof,
			&b.EncryptedBalanceData,
			&encryptedTxs,
			&encryptedTransfers,
			&encryptedDeposits,
			&b.Signature,
			&b.BlockNumber,
			&b.CreatedAt,
		)
		if err != nil {
			return err
		}

		err = unmarshalBackupBalanceData(&b, encryptedTxs, encryptedTransfers, encryptedDeposits)
		if err != nil {
			return err
		}

		balance := p.backupBalanceToDBApp(&b)

		balances = append(balances, &balance)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return balances, nil
}

func (p *pgx) backupBalanceToDBApp(b *models.BackupBalance) mDBApp.BackupBalance {
	return mDBApp.BackupBalance{
		ID:                    b.ID,
		UserAddress:           b.UserAddress,
		EncryptedBalanceProof: b.EncryptedBalanceProof,
		EncryptedBalanceData:  b.EncryptedBalanceData,
		EncryptedTxs:          b.EncryptedTxs,
		EncryptedTransfers:    b.EncryptedTransfers,
		EncryptedDeposits:     b.EncryptedDeposits,
		BlockNumber:           b.BlockNumber,
		CreatedAt:             b.CreatedAt,
	}
}

func unmarshalBackupBalanceData(
	b *models.BackupBalance,
	encryptedTxs, encryptedTransfers, encryptedDeposits []byte,
) (err error) {
	err = json.Unmarshal(encryptedTxs, &b.EncryptedTxs)
	if err != nil {
		const msg = "failed to unmarshal EncryptedTxs: %w"
		return fmt.Errorf(msg, err)
	}

	err = json.Unmarshal(encryptedTransfers, &b.EncryptedTransfers)
	if err != nil {
		const msg = "failed to unmarshal EncryptedTransfers: %w"
		return fmt.Errorf(msg, err)
	}

	err = json.Unmarshal(encryptedDeposits, &b.EncryptedDeposits)
	if err != nil {
		const msg = "failed to unmarshal EncryptedDeposits: %w"
		return fmt.Errorf(msg, err)
	}

	return nil
}
