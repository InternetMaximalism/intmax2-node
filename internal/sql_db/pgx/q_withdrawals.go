package pgx

import (
	"encoding/json"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"log"
	"time"
)

func (p *pgx) CreateWithdrawal(id string, input postWithdrwalRequest.UCPostWithdrawalRequestInput) (*mDBApp.Withdrawal, error) {
	const (
		query = ` INSERT INTO withdrawals
              (id, status, transfer_data, transfer_merkle_proof, transaction, tx_merkle_proof, enough_balance_proof, transfer_hash, block_number, block_hash, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) `
	)

	status := mDBApp.PENDING
	transferHash := input.TransferHash
	blockNumber := input.BlockNumber
	blockHash := input.BlockHash
	createdAt := time.Now().UTC()

	transferDataJSON, err := json.Marshal(input.TransferData)
	if err != nil {
		log.Fatalf("Error encoding TransferData: %v", err)
	}
	transferMerkleProofJSON, err := json.Marshal(input.TransferMerkleProof)
	if err != nil {
		log.Fatalf("Error encoding TransferMerkleProof: %v", err)
	}
	transactionJSON, err := json.Marshal(input.Transaction)
	if err != nil {
		log.Fatalf("Error encoding Transaction: %v", err)
	}
	txMerkleProofJSON, err := json.Marshal(input.TxMerkleProof)
	if err != nil {
		log.Fatalf("Error encoding TxMerkleProof: %v", err)
	}
	enoughBalanceProofJSON, err := json.Marshal(input.EnoughBalanceProof)
	if err != nil {
		log.Fatalf("Error encoding EnoughBalanceProof: %v", err)
	}

	_, err = p.exec(
		p.ctx,
		query,
		id,
		status,
		transferDataJSON,
		transferMerkleProofJSON,
		transactionJSON,
		txMerkleProofJSON,
		enoughBalanceProofJSON,
		transferHash,
		blockNumber,
		blockHash,
		createdAt,
	)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var wDBApp *mDBApp.Withdrawal
	wDBApp, err = p.WithdrawalByID(id)
	if err != nil {
		return nil, err
	}

	return wDBApp, nil
}

func (p *pgx) WithdrawalByID(id string) (*mDBApp.Withdrawal, error) {
	const q = `
	    SELECT id, status, transfer_data, transfer_merkle_proof, transaction, tx_merkle_proof, enough_balance_proof, transfer_hash, block_number, block_hash, created_at
	    FROM withdrawals
        WHERE id = $1
    `

	var tmp models.Withdrawal
	err := errPgx.Err(p.queryRow(p.ctx, q, id).
		Scan(
			&tmp.ID,
			&tmp.Status,
			&tmp.TransferData,
			&tmp.TransferMerkleProof,
			&tmp.Transaction,
			&tmp.TxMerkleProof,
			&tmp.EnoughBalanceProof,
			&tmp.TransferHash,
			&tmp.BlockNumber,
			&tmp.BlockHash,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	wDBApp := p.wToDBApp(&tmp)

	return &wDBApp, nil
}

func (p *pgx) WithdrawalsByStatus(status mDBApp.WithdrawalStatus) (*[]mDBApp.Withdrawal, error) {
	const q = `
	    SELECT id, status, transfer_data, transfer_merkle_proof, transaction, tx_merkle_proof, enough_balance_proof, transfer_hash, block_number, block_hash, created_at
	    FROM withdrawals
	    WHERE status = $1
    `

	rows, err := p.query(p.ctx, q, status)
	if err != nil {
		return nil, errPgx.Err(err)
	}
	defer rows.Close()

	var withdrawals []mDBApp.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		err := rows.Scan(
			&w.ID,
			&w.Status,
			&w.TransferData,
			&w.TransferMerkleProof,
			&w.Transaction,
			&w.TxMerkleProof,
			&w.EnoughBalanceProof,
			&w.TransferHash,
			&w.BlockNumber,
			&w.BlockHash,
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

func (p *pgx) wToDBApp(w *models.Withdrawal) mDBApp.Withdrawal {
	m := mDBApp.Withdrawal{
		ID: w.ID,
		TransferData: mDBApp.TransferData{
			Recipient:  w.TransferData.Recipient,
			TokenIndex: w.TransferData.TokenIndex,
			Amount:     w.TransferData.Amount,
			Salt:       w.TransferData.Salt,
		},
		TransferMerkleProof: mDBApp.TransferMerkleProof{
			Index:    w.TransferMerkleProof.Index,
			Siblings: w.TransferMerkleProof.Siblings,
		},
		Transaction: mDBApp.Transaction{
			TransferTreeRoot: w.Transaction.TransferTreeRoot,
			Nonce:            w.Transaction.Nonce,
		},
		TxMerkleProof: mDBApp.TxMerkleProof{
			Index:    w.TxMerkleProof.Index,
			Siblings: w.TxMerkleProof.Siblings,
		},
		EnoughBalanceProof: mDBApp.EnoughBalanceProof{
			Proof:        w.EnoughBalanceProof.Proof,
			PublicInputs: w.EnoughBalanceProof.PublicInputs,
		},
		TransferHash: w.TransferHash,
		BlockNumber:  w.BlockNumber,
		CreatedAt:    w.CreatedAt,
	}

	return m
}
