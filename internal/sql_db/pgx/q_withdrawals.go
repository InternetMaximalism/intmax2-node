package pgx

import (
	"encoding/json"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"log"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateWithdrawal(w *mDBApp.Withdrawal) (*mDBApp.Withdrawal, error) {
	const (
		query = ` INSERT INTO withdrawals
              (id, status, recipient, token_index, amount, salt, transfer_hash, transfer_merkle_proof, transaction, tx_merkle_proof, block_number, enough_balance_proof, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) `
	)

	id := uuid.New().String()
	status := mDBApp.PENDING
	recipient := w.Recipient
	tokenIndex := w.TokenIndex
	amount := w.Amount
	salt := w.Salt
	transferHash := w.TransferHash
	blockNumber := w.BlockNumber
	enoughBalanceProof := w.EnoughBalanceProof
	createdAt := time.Now().UTC()

	transferMerkleProofJSON, err := json.Marshal(w.TransferMerkleProof)
	if err != nil {
		log.Fatalf("Error encoding TransferMerkleProof: %v", err)
	}
	transactionJSON, err := json.Marshal(w.Transaction)
	if err != nil {
		log.Fatalf("Error encoding Transaction: %v", err)
	}
	txMerkleProofJSON, err := json.Marshal(w.TxMerkleProof)
	if err != nil {
		log.Fatalf("Error encoding TxMerkleProof: %v", err)
	}

	_, err = p.exec(
		p.ctx,
		query,
		id,
		status,
		recipient,
		tokenIndex,
		amount,
		salt,
		transferHash,
		transferMerkleProofJSON,
		transactionJSON,
		txMerkleProofJSON,
		blockNumber,
		enoughBalanceProof,
		createdAt,
	)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var tDBApp *mDBApp.Withdrawal
	// tDBApp, err = p.TokenByIndex(tokenIndex)
	// if err != nil {
	// 	return nil, err
	// }

	return tDBApp, nil
}

func (p *pgx) FindWithdrawals(status mDBApp.WithdrawalStatus) (*[]mDBApp.Withdrawal, error) {
	const (
		q = ` SELECT id, status, created_at FROM withdrawals WHERE status = $1 `
	)

	rows, err := p.query(p.ctx, q, status)
	if err != nil {
		return nil, errPgx.Err(err)
	}
	defer rows.Close()

	var withdrawals []mDBApp.Withdrawal
	for rows.Next() {
		var w mDBApp.Withdrawal
		err := rows.Scan(
			&w.ID,
			&w.Status,
			&w.CreatedAt,
		)
		if err != nil {
			return nil, errPgx.Err(err)
		}
		withdrawals = append(withdrawals, p.wToDBApp(&w))
	}

	if rows.Err() != nil {
		return nil, errPgx.Err(rows.Err())
	}

	return &withdrawals, nil
}

func (p *pgx) wToDBApp(w *mDBApp.Withdrawal) mDBApp.Withdrawal {
	m := mDBApp.Withdrawal{
		ID:           w.ID,
		Recipient:    w.Recipient,
		TokenIndex:   w.TokenIndex,
		Amount:       w.Amount,
		Salt:         w.Salt,
		TransferHash: w.TransferHash,
		TransferMerkleProof: mDBApp.TransferMerkleProof{
			Index:    w.TransferMerkleProof.Index,
			Siblings: w.TransferMerkleProof.Siblings,
		},
		Transaction: mDBApp.Transaction{
			FeeTransferHash:  w.Transaction.FeeTransferHash,
			TransferTreeRoot: w.Transaction.TransferTreeRoot,
			TokenIndex:       w.Transaction.TokenIndex,
			Nonce:            w.Transaction.Nonce,
		},
		TxMerkleProof: mDBApp.TxMerkleProof{
			Index:    w.TxMerkleProof.Index,
			Siblings: w.TxMerkleProof.Siblings,
		},
		BlockNumber:        w.BlockNumber,
		EnoughBalanceProof: w.EnoughBalanceProof,
		CreatedAt:          w.CreatedAt,
	}

	return m
}
