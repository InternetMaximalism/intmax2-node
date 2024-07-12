package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateTransaction(
	senderPublicKey, txHash, signatureID string,
) (*mDBApp.Transactions, error) {
	tx := models.Transactions{
		TxID:            uuid.New().String(),
		TxHash:          txHash,
		SenderPublicKey: senderPublicKey,
		SignatureID:     signatureID,
		CreatedAt:       time.Now().UTC(),
	}

	const (
		q = ` INSERT INTO transactions
              (tx_id, tx_hash, sender_public_key, signature_id, created_at)
              VALUES ($1, $2, $3, $4, $5) `
	)

	_, err := p.exec(p.ctx, q,
		tx.TxID, tx.TxHash, tx.SenderPublicKey, tx.SignatureID, tx.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var tDBApp *mDBApp.Transactions
	tDBApp, err = p.TransactionByID(tx.TxID)
	if err != nil {
		return nil, err
	}

	return tDBApp, nil
}

func (p *pgx) TransactionByID(txID string) (*mDBApp.Transactions, error) {
	const (
		q = `SELECT tx_id, tx_hash, sender_public_key, signature_id, status, created_at
             FROM transactions WHERE tx_id = $1`
	)

	var tmp models.Transactions
	err := errPgx.Err(p.queryRow(p.ctx, q, txID).
		Scan(
			&tmp.TxID,
			&tmp.TxHash,
			&tmp.SenderPublicKey,
			&tmp.SignatureID,
			&tmp.Status,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	txDBApp := p.txToDBApp(&tmp)

	return &txDBApp, nil
}

func (p *pgx) txToDBApp(tx *models.Transactions) mDBApp.Transactions {
	m := mDBApp.Transactions{
		TxID:            tx.TxID,
		TxHash:          tx.TxHash,
		SenderPublicKey: tx.SenderPublicKey,
		SignatureID:     tx.SignatureID,
		CreatedAt:       tx.CreatedAt,
	}
	if tx.Status.Valid {
		m.Status = tx.Status.Int64
	}

	return m
}
