package pgx

import (
	"database/sql"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupTransfer(
	recipient, encryptedTransferHash, encryptedTransfer string,
	blockNumber int64,
) (*mDBApp.BackupTransfer, error) {
	const query = `
	    INSERT INTO backup_transfers
        (id, recipient, transfer_double_hash, encrypted_transfer, block_number, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query, id, recipient, encryptedTransferHash, encryptedTransfer, blockNumber, createdAt)
	if err != nil {
		return nil, err
	}

	return p.GetBackupTransfer("id", id)
}

func (p *pgx) GetBackupTransfer(condition, value string) (*mDBApp.BackupTransfer, error) {
	const baseQuery = `
        SELECT id, recipient, transfer_double_hash, encrypted_transfer, block_number, created_at
        FROM backup_transfers
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)

	var b models.BackupTransfer
	err := errPgx.Err(p.queryRow(p.ctx, query, value).
		Scan(
			&b.ID,
			&b.Recipient,
			&b.TransferDoubleHash,
			&b.EncryptedTransfer,
			&b.BlockNumber,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}
	transfer := p.backupTransferToDBApp(&b)
	return &transfer, nil
}

func (p *pgx) GetBackupTransfers(condition string, value interface{}) ([]*mDBApp.BackupTransfer, error) {
	const baseQuery = `
        SELECT id, recipient, transfer_double_hash, encrypted_transfer, block_number, created_at
        FROM backup_transfers
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)
	var transfers []*mDBApp.BackupTransfer
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupTransfer
		err := rows.Scan(&b.ID, &b.Recipient, &b.TransferDoubleHash, &b.EncryptedTransfer, &b.BlockNumber, &b.CreatedAt)
		if err != nil {
			return err
		}
		transfer := p.backupTransferToDBApp(&b)
		transfers = append(transfers, &transfer)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transfers, nil
}

func (p *pgx) backupTransferToDBApp(b *models.BackupTransfer) mDBApp.BackupTransfer {
	return mDBApp.BackupTransfer{
		ID:                 b.ID,
		Recipient:          b.Recipient,
		TransferDoubleHash: b.TransferDoubleHash.String,
		EncryptedTransfer:  b.EncryptedTransfer,
		BlockNumber:        b.BlockNumber,
		CreatedAt:          b.CreatedAt,
	}
}
