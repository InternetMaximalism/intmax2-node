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

func (p *pgx) CreateBackupTransaction(
	sender, encryptedTxHash, encryptedTx, signature string,
	blockNumber int64,
) (*mDBApp.BackupTransaction, error) {
	const query = `
	    INSERT INTO backup_transactions
        (id, sender, tx_double_hash, encrypted_tx, block_number, signature, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	id := uuid.New().String()
	createdAt := time.Now().UTC()

	err := p.createBackupEntry(query,
		id, sender, encryptedTxHash, encryptedTx, blockNumber, signature, createdAt,
	)
	if err != nil {
		return nil, err
	}

	return p.GetBackupTransaction("id", id)
}

func (p *pgx) GetBackupTransaction(condition, value string) (*mDBApp.BackupTransaction, error) {
	const baseQuery = `
        SELECT id, sender, tx_double_hash, encrypted_tx, block_number, signature, created_at
        FROM backup_transactions
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)

	var b models.BackupTransaction
	err := errPgx.Err(p.queryRow(p.ctx, query, value).
		Scan(
			&b.ID,
			&b.Sender,
			&b.TxDoubleHash,
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

func (p *pgx) GetBackupTransactionBySenderAndTxDoubleHash(sender, txDoubleHash string) (*mDBApp.BackupTransaction, error) {
	const (
		q = `
        SELECT id, sender, tx_double_hash, encrypted_tx, block_number, signature, created_at
        FROM backup_transactions
        WHERE sender = $1 AND tx_double_hash = $2 `
	)

	var b models.BackupTransaction
	err := errPgx.Err(p.queryRow(p.ctx, q, sender, txDoubleHash).
		Scan(
			&b.ID,
			&b.Sender,
			&b.TxDoubleHash,
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
        SELECT id, sender, tx_double_hash, encrypted_tx, block_number, signature, created_at
        FROM backup_transactions
        WHERE %s = $1
`
	query := fmt.Sprintf(baseQuery, condition)
	var transactions []*mDBApp.BackupTransaction
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupTransaction
		err := rows.Scan(&b.ID, &b.Sender, &b.TxDoubleHash, &b.EncryptedTx, &b.BlockNumber, &b.Signature, &b.CreatedAt)
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
