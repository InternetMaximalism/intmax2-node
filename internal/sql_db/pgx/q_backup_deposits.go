package pgx

import (
	"database/sql"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupDeposit(
	recipient, depositHash, encryptedDeposit string,
	blockNumber int64,
) (*mDBApp.BackupDeposit, error) {
	const query = `
	    INSERT INTO backup_deposits
        (id, recipient, deposit_double_hash, encrypted_deposit, block_number, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query, id, recipient, depositHash, encryptedDeposit, blockNumber, createdAt)
	if err != nil {
		return nil, err
	}

	return p.GetBackupDeposit([]string{"id"}, []interface{}{id})
}

func (p *pgx) GetBackupDeposit(conditions []string, values []interface{}) (*mDBApp.BackupDeposit, error) {
	const baseQuery = `
        SELECT id, recipient, deposit_double_hash, encrypted_deposit, block_number, created_at 
        FROM backup_deposits 
        WHERE %s`

	whereClause := make([]string, len(conditions))
	for i, condition := range conditions {
		whereClause[i] = fmt.Sprintf("%s = $%d", condition, i+1)
	}

	query := fmt.Sprintf(baseQuery, strings.Join(whereClause, " AND "))

	var b models.BackupDeposit
	err := errPgx.Err(p.queryRow(p.ctx, query, values...).
		Scan(
			&b.ID,
			&b.Recipient,
			&b.DepositDoubleHash,
			&b.EncryptedDeposit,
			&b.BlockNumber,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	deposit := p.backupDepositToDBApp(&b)
	return &deposit, nil
}

func (p *pgx) GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error) {
	const baseQuery = `
        SELECT id, recipient, deposit_double_hash, encrypted_deposit, block_number, created_at
        FROM backup_deposits
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)
	var deposits []*mDBApp.BackupDeposit
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupDeposit
		err := rows.Scan(&b.ID, &b.Recipient, &b.DepositDoubleHash, &b.EncryptedDeposit, &b.BlockNumber, &b.CreatedAt)
		if err != nil {
			return err
		}
		deposit := p.backupDepositToDBApp(&b)
		deposits = append(deposits, &deposit)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return deposits, nil
}

func (p *pgx) backupDepositToDBApp(b *models.BackupDeposit) mDBApp.BackupDeposit {
	return mDBApp.BackupDeposit{
		ID:                b.ID,
		Recipient:         b.Recipient,
		DepositDoubleHash: b.DepositDoubleHash.String,
		EncryptedDeposit:  b.EncryptedDeposit,
		BlockNumber:       b.BlockNumber,
		CreatedAt:         b.CreatedAt,
	}
}
