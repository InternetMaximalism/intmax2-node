package pgx

import (
	"encoding/json"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
	"github.com/holiman/uint256"
)

func (p *pgx) CreateTxMerkleProofs(
	senderPublicKey, txHash, txID string,
	txTreeIndex *uint256.Int,
	txMerkleProof json.RawMessage,
) (*mDBApp.TxMerkleProofs, error) {
	tmp := models.TxMerkleProofs{
		ID:              uuid.New().String(),
		SenderPublicKey: senderPublicKey,
		TxHash:          txHash,
		TxTreeIndex:     txTreeIndex,
		TxMerkleProof:   txMerkleProof,
		CreatedAt:       time.Now().UTC(),
	}

	txiID, _ := tmp.TxTreeIndex.Value()

	const (
		q = ` INSERT INTO tx_merkle_proofs
              (id, sender_public_key, tx_hash, tx_id, tx_tree_index, tx_merkle_proof, created_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7) `
	)

	_, err := p.exec(p.ctx, q,
		tmp.ID, tmp.SenderPublicKey, tmp.TxHash, txID, txiID, tmp.TxMerkleProof, tmp.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var tDBApp *mDBApp.TxMerkleProofs
	tDBApp, err = p.TxMerkleProofsByID(tmp.ID)
	if err != nil {
		return nil, err
	}

	return tDBApp, nil
}

func (p *pgx) TxMerkleProofsByID(id string) (*mDBApp.TxMerkleProofs, error) {
	const (
		q = ` SELECT id, sender_public_key, tx_hash, tx_id, tx_tree_index, tx_merkle_proof, created_at
              FROM tx_merkle_proofs WHERE id = $1 `
	)

	var tmp models.TxMerkleProofs
	err := errPgx.Err(p.queryRow(p.ctx, q, id).
		Scan(
			&tmp.ID,
			&tmp.SenderPublicKey,
			&tmp.TxHash,
			&tmp.TxID,
			&tmp.TxTreeIndex,
			&tmp.TxMerkleProof,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	tmpDBApp := p.tmpToDBApp(&tmp)

	return &tmpDBApp, nil
}

func (p *pgx) TxMerkleProofsByTxHash(txHash string) (*mDBApp.TxMerkleProofs, error) {
	const (
		q = ` SELECT id, sender_public_key, tx_hash, tx_id, tx_tree_index, tx_merkle_proof, created_at
              FROM tx_merkle_proofs WHERE tx_hash = $1 `
	)

	var tmp models.TxMerkleProofs
	err := errPgx.Err(p.queryRow(p.ctx, q, txHash).
		Scan(
			&tmp.ID,
			&tmp.SenderPublicKey,
			&tmp.TxHash,
			&tmp.TxID,
			&tmp.TxTreeIndex,
			&tmp.TxMerkleProof,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	tmpDBApp := p.tmpToDBApp(&tmp)

	return &tmpDBApp, nil
}

func (p *pgx) tmpToDBApp(tmp *models.TxMerkleProofs) mDBApp.TxMerkleProofs {
	m := mDBApp.TxMerkleProofs{
		ID:              tmp.ID,
		SenderPublicKey: tmp.SenderPublicKey,
		TxHash:          tmp.TxHash,
		TxID:            tmp.TxID,
		TxTreeIndex:     tmp.TxTreeIndex,
		TxMerkleProof:   tmp.TxMerkleProof,
		CreatedAt:       tmp.CreatedAt,
	}

	return m
}
