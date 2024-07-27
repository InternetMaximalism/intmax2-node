package pgx

import (
	"encoding/json"
	"fmt"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"
)

func (p *pgx) CreateWithdrawal(id string, input postWithdrwalRequest.UCPostWithdrawalRequestInput) (*mDBApp.Withdrawal, error) {
	const query = `
	    INSERT INTO withdrawals
        (id, status, transfer_data, transfer_merkle_proof, transaction, tx_merkle_proof, enough_balance_proof, transfer_hash, block_number, block_hash, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	createdAt := time.Now().UTC()

	jsonFields := map[string]interface{}{
		"TransferData":        input.TransferData,
		"TransferMerkleProof": input.TransferMerkleProof,
		"Transaction":         input.Transaction,
		"TxMerkleProof":       input.TxMerkleProof,
		"EnoughBalanceProof":  input.EnoughBalanceProof,
	}

	jsonData := make(map[string][]byte)
	for field, data := range jsonFields {
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("error encoding %s: %w", field, err)
		}
		jsonData[field] = jsonBytes
	}

	_, err := p.exec(
		p.ctx,
		query,
		id,
		mDBApp.WS_PENDING,
		jsonData["TransferData"],
		jsonData["TransferMerkleProof"],
		jsonData["Transaction"],
		jsonData["TxMerkleProof"],
		jsonData["EnoughBalanceProof"],
		input.TransferHash,
		input.BlockNumber,
		input.BlockHash,
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

func (p *pgx) UpdateWithdrawalsStatus(ids []string, status mDBApp.WithdrawalStatus) error {
	placeholder := make([]string, len(ids))
	for i := range ids {
		placeholder[i] = fmt.Sprintf("$%d", i+1)
	}
	placeholderStr := strings.Join(placeholder, ", ")

	query := fmt.Sprintf(`
	    UPDATE withdrawals
		SET status = %d
		WHERE id IN (%s)
	`, status, placeholderStr)

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	_, err := p.exec(p.ctx, query, args...)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) WithdrawalByID(id string) (*mDBApp.Withdrawal, error) {
	const query = `
	    SELECT id, status, transfer_data, transfer_merkle_proof, transaction, tx_merkle_proof, enough_balance_proof, transfer_hash, block_number, block_hash, created_at
	    FROM withdrawals
        WHERE id = $1
    `

	var tmp models.Withdrawal
	var transferDataJSON, transferMerkleProofJSON, transactionJSON, txMerkleProofJSON, enoughBalanceProofJSON []byte
	err := errPgx.Err(p.queryRow(p.ctx, query, id).
		Scan(
			&tmp.ID,
			&tmp.Status,
			&transferDataJSON,
			&transferMerkleProofJSON,
			&transactionJSON,
			&txMerkleProofJSON,
			&enoughBalanceProofJSON,
			&tmp.TransferHash,
			&tmp.BlockNumber,
			&tmp.BlockHash,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	if err := unmarshalWithdrawalData(&tmp, transferDataJSON, transferMerkleProofJSON, transactionJSON, txMerkleProofJSON, enoughBalanceProofJSON); err != nil {
		return nil, err
	}

	wDBApp := p.wToDBApp(&tmp)

	return &wDBApp, nil
}

func (p *pgx) WithdrawalsByStatus(status mDBApp.WithdrawalStatus, limit *int) (*[]mDBApp.Withdrawal, error) {
	baseQuery := `
        SELECT id, status, transfer_data, transfer_merkle_proof, transaction, tx_merkle_proof, enough_balance_proof, transfer_hash, block_number, block_hash, created_at
        FROM withdrawals
        WHERE status = $1
        ORDER BY created_at ASC`

	var query string
	var args []interface{}
	args = append(args, status)

	if limit != nil {
		query = baseQuery + " LIMIT $2"
		args = append(args, *limit)
	} else {
		query = baseQuery
	}

	rows, err := p.query(p.ctx, query, args...)
	if err != nil {
		return nil, errPgx.Err(err)
	}
	defer rows.Close()

	var withdrawals []mDBApp.Withdrawal
	for rows.Next() {
		var w models.Withdrawal
		var transferDataJSON, transferMerkleProofJSON, transactionJSON, txMerkleProofJSON, enoughBalanceProofJSON []byte

		err := rows.Scan(
			&w.ID, &w.Status, &transferDataJSON, &transferMerkleProofJSON,
			&transactionJSON, &txMerkleProofJSON, &enoughBalanceProofJSON,
			&w.TransferHash, &w.BlockNumber, &w.BlockHash, &w.CreatedAt,
		)
		if err != nil {
			return nil, errPgx.Err(err)
		}

		if err := unmarshalWithdrawalData(&w, transferDataJSON, transferMerkleProofJSON, transactionJSON, txMerkleProofJSON, enoughBalanceProofJSON); err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, p.wToDBApp(&w))
	}

	if err := rows.Err(); err != nil {
		return nil, errPgx.Err(err)
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

func unmarshalWithdrawalData(w *models.Withdrawal, transferDataJSON, transferMerkleProofJSON, transactionJSON, txMerkleProofJSON, enoughBalanceProofJSON []byte) error {
	if err := json.Unmarshal(transferDataJSON, &w.TransferData); err != nil {
		return fmt.Errorf("failed to unmarshal TransferData: %w", err)
	}
	if err := json.Unmarshal(transferMerkleProofJSON, &w.TransferMerkleProof); err != nil {
		return fmt.Errorf("failed to unmarshal TransferMerkleProof: %w", err)
	}
	if err := json.Unmarshal(transactionJSON, &w.Transaction); err != nil {
		return fmt.Errorf("failed to unmarshal Transaction: %w", err)
	}
	if err := json.Unmarshal(txMerkleProofJSON, &w.TxMerkleProof); err != nil {
		return fmt.Errorf("failed to unmarshal TxMerkleProof: %w", err)
	}
	if err := json.Unmarshal(enoughBalanceProofJSON, &w.EnoughBalanceProof); err != nil {
		return fmt.Errorf("failed to unmarshal EnoughBalanceProof: %w", err)
	}
	return nil
}
