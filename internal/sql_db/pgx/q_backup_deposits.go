package pgx

import (
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

	_, err := p.exec(
		p.ctx,
		query,
		id,
		input.Recipient,
		input.EncryptedDeposit,
		input.BlockNumber,
		createdAt,
	)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var mDBApp *mDBApp.BackupDeposit
	mDBApp, err = p.GetBackupDeposit("id", id)
	if err != nil {
		return nil, err
	}

	return mDBApp, nil
}

func (p *pgx) GetBackupDeposit(condition string, value string) (*mDBApp.BackupDeposit, error) {
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
	mDBApp := p.backupDepositToDBApp(&b)
	return &mDBApp, nil
}

func (p *pgx) GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error) {
	const baseQuery = `
        SELECT id, recipient, encrypted_deposit, block_number, created_at
        FROM backup_deposits
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)

	rows, err := p.query(p.ctx, query, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deposits []*mDBApp.BackupDeposit

	for rows.Next() {
		var b models.BackupDeposit
		err := rows.Scan(
			&b.ID,
			&b.Recipient,
			&b.EncryptedDeposit,
			&b.BlockNumber,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		mDBApp := p.backupDepositToDBApp(&b)
		deposits = append(deposits, &mDBApp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return deposits, nil
}

func (p *pgx) backupDepositToDBApp(b *models.BackupDeposit) mDBApp.BackupDeposit {
	m := mDBApp.BackupDeposit{
		ID:               b.ID,
		Recipient:        b.Recipient,
		EncryptedDeposit: b.EncryptedDeposit,
		BlockNumber:      b.BlockNumber,
		CreatedAt:        b.CreatedAt,
	}

	return m
}
