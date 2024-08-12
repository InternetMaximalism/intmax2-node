package pgx

import (
	"database/sql"
	"encoding/json"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBackupBalance(input *backupBalance.UCPostBackupBalanceInput) (*mDBApp.BackupBalance, error) {
	const query = `
	    INSERT INTO backup_balances
        (id, user_address, encrypted_balance_proof, encrypted_balance_data, encrypted_txs, encrypted_transfers, encrypted_deposits, signature, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	id := uuid.New().String()
	createdAt := time.Now().UTC()

	encryptedTxs, err := json.Marshal(input.EncryptedTxs)
	if err != nil {
		return nil, fmt.Errorf("error encoding EncryptedTxs: %w", err)
	}
	encryptedTransfers, err := json.Marshal(input.EncryptedTransfers)
	if err != nil {
		return nil, fmt.Errorf("error encoding EncryptedTransfers: %w", err)
	}
	encryptedDeposits, err := json.Marshal(input.EncryptedDeposits)
	if err != nil {
		return nil, fmt.Errorf("error encoding EncryptedDeposits: %w", err)
	}

	_, err = p.db.Exec(query,
		id,
		input.User,
		input.EncryptedBalanceProof,
		input.EncryptedBalanceData,
		encryptedTxs,
		encryptedTransfers,
		encryptedDeposits,
		input.Signature,
		createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create backup balance: %w", err)
	}

	return p.GetBackupBalance([]string{"id"}, []interface{}{id})
}

func (p *pgx) GetBackupBalance(conditions []string, values []interface{}) (*mDBApp.BackupBalance, error) {
	const baseQuery = `
        SELECT id, user_address, encrypted_balance_proof, encrypted_balance_data, encrypted_txs, encrypted_transfers, encrypted_deposits, signature, created_at
        FROM backup_balances 
        WHERE %s`

	whereClause := make([]string, len(conditions))
	for i, condition := range conditions {
		whereClause[i] = fmt.Sprintf("%s = $%d", condition, i+1)
	}
	query := fmt.Sprintf(baseQuery, strings.Join(whereClause, " AND "))

	var b models.BackupBalance
	var encryptedTxs, encryptedTransfers, encryptedDeposits []byte
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
	const baseQuery = `
        SELECT id, user_address, encrypted_balance_proof, encrypted_balance_data, encrypted_txs, encrypted_transfers, encrypted_deposits, signature, created_at
        FROM backup_balances
        WHERE %s = $1
    `
	query := fmt.Sprintf(baseQuery, condition)
	var balances []*mDBApp.BackupBalance
	err := p.getBackupEntries(query, value, func(rows *sql.Rows) error {
		var b models.BackupBalance
		err := rows.Scan(&b.ID, &b.UserAddress, &b.EncryptedBalanceProof, &b.EncryptedBalanceData, &b.EncryptedTxs, &b.EncryptedTransfers, &b.EncryptedDeposits, &b.Signature, &b.CreatedAt)
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
		CreatedAt:             b.CreatedAt,
	}
}

func unmarshalBackupBalanceData(b *models.BackupBalance, encryptedTxs, encryptedTransfers, encryptedDeposits []byte) error {
	var err error
	if err = json.Unmarshal(encryptedTxs, &b.EncryptedTxs); err != nil {
		return fmt.Errorf("failed to unmarshal EncryptedTxs: %w", err)
	}
	if err = json.Unmarshal(encryptedTransfers, &b.EncryptedTransfers); err != nil {
		return fmt.Errorf("failed to unmarshal EncryptedTransfers: %w", err)
	}
	if err = json.Unmarshal(encryptedDeposits, &b.EncryptedDeposits); err != nil {
		return fmt.Errorf("failed to unmarshal EncryptedDeposits: %w", err)
	}
	return nil
}
