package pgx

import (
	"database/sql"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupTransaction(input *backupTransaction.UCPostBackupTransactionInput) (*mDBApp.BackupTransaction, error) {
	const query = `
	    INSERT INTO backup_transactions
        (id, sender, encrypted_tx, block_number, signature, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query, id, input.Sender, input.EncryptedTx, input.BlockNumber, input.Signature, createdAt)
	if err != nil {
		return nil, err
	}

	return p.GetBackupTransaction("id", id)
}

func (p *pgx) GetBackupTransaction(condition, value string) (*mDBApp.BackupTransaction, error) {
	const baseQuery = `
        SELECT id, sender, encrypted_tx, block_number, signature, created_at
        FROM backup_transactions
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)

	var b models.BackupTransaction
	err := errPgx.Err(p.queryRow(p.ctx, query, value).
		Scan(
			&b.ID,
			&b.Sender,
			&b.EncryptedTx,
			&b.BlockNumber,
			&b.Signature,
			&b.CreatedAt,
		))
	if err != nil {
		return nil, err
	}
	transaction := p.backupTransactionToDBApp(&b)
	return &transaction, nil
}

func (p *pgx) GetBackupTransactions(condition string, value interface{}) ([]*mDBApp.BackupTransaction, error) {
	const baseQuery = `
        SELECT id, sender, encrypted_tx, block_number, signature, created_at
        FROM backup_transactions
        WHERE %s = $1
`
	query := fmt.Sprintf(baseQuery, condition)
	var transactions []*mDBApp.BackupTransaction
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupTransaction
		err := rows.Scan(&b.ID, &b.Sender, &b.EncryptedTx, &b.BlockNumber, &b.Signature, &b.CreatedAt)
		if err != nil {
			return err
		}
		transaction := p.backupTransactionToDBApp(&b)
		transactions = append(transactions, &transaction)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (p *pgx) backupTransactionToDBApp(b *models.BackupTransaction) mDBApp.BackupTransaction {
	return mDBApp.BackupTransaction{
		ID:           b.ID,
		Sender:       b.Sender,
		TxDoubleHash: b.TxDoubleHash.String,
		EncryptedTx:  b.EncryptedTx,
		BlockNumber:  b.BlockNumber,
		Signature:    b.Signature,
		CreatedAt:    b.CreatedAt,
	}
}
