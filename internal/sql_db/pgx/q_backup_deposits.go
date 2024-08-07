package pgx

import (
	"database/sql"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupDeposit(input *backupDeposit.UCPostBackupDepositInput) (*mDBApp.BackupDeposit, error) {
	const query = `
	    INSERT INTO backup_deposits
        (id, recipient, encrypted_deposit, block_number, created_at)
        VALUES ($1, $2, $3, $4, $5)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query, id, input.Recipient, input.EncryptedDeposit, input.BlockNumber, createdAt)
	if err != nil {
		return nil, err
	}

	return p.GetBackupDeposit("id", id)
}

func (p *pgx) GetBackupDeposit(condition, value string) (*mDBApp.BackupDeposit, error) {
	const baseQuery = `
        SELECT id, recipient, encrypted_deposit, block_number, created_at
        FROM backup_deposits
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)

	var b models.BackupDeposit
	err := errPgx.Err(p.queryRow(p.ctx, query, value).
		Scan(
			&b.ID,
			&b.Recipient,
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
        SELECT id, recipient, encrypted_deposit, block_number, created_at
        FROM backup_deposits
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)
	var deposits []*mDBApp.BackupDeposit
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupDeposit
		err := rows.Scan(&b.ID, &b.Recipient, &b.EncryptedDeposit, &b.BlockNumber, &b.CreatedAt)
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
		ID:               b.ID,
		Recipient:        b.Recipient,
		EncryptedDeposit: b.EncryptedDeposit,
		BlockNumber:      b.BlockNumber,
		CreatedAt:        b.CreatedAt,
	}
}
