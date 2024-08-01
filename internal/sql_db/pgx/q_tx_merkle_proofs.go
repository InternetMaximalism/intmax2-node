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
	senderPublicKey, txHash, signatureID string,
	txTreeIndex *uint256.Int,
	txMerkleProof json.RawMessage,
	txTreeRoot string,
	proposalBlockID string,
) (*mDBApp.TxMerkleProofs, error) {
	tmp := models.TxMerkleProofs{
		ID:              uuid.New().String(),
		SenderPublicKey: senderPublicKey,
		TxHash:          txHash,
		TxTreeIndex:     txTreeIndex,
		TxMerkleProof:   txMerkleProof,
		TxTreeRoot:      txTreeRoot,
		ProposalBlockID: proposalBlockID,
		CreatedAt:       time.Now().UTC(),
	}

	txiID, _ := tmp.TxTreeIndex.Value()

	const (
		qWithoutSign = ` INSERT INTO tx_merkle_proofs (
              id ,sender_public_key ,tx_hash ,tx_tree_index ,tx_merkle_proof
              ,tx_tree_root ,proposal_block_id ,created_at
              ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) `

		qWithSign = ` INSERT INTO tx_merkle_proofs (
              id ,sender_public_key ,tx_hash ,tx_tree_index ,tx_merkle_proof
              ,tx_tree_root ,signature_id ,proposal_block_id ,created_at
              ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) `
	)

	var err error
	if signatureID == "" {
		_, err = p.exec(p.ctx, qWithoutSign,
			tmp.ID, tmp.SenderPublicKey, tmp.TxHash, txiID,
			tmp.TxMerkleProof, tmp.TxTreeRoot, tmp.ProposalBlockID, tmp.CreatedAt,
		)
	} else {
		_, err = p.exec(p.ctx, qWithSign,
			tmp.ID, tmp.SenderPublicKey, tmp.TxHash, txiID,
			tmp.TxMerkleProof, tmp.TxTreeRoot, signatureID, tmp.ProposalBlockID, tmp.CreatedAt,
		)
	}
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
		q = ` SELECT
              id ,sender_public_key ,tx_hash ,tx_tree_index ,tx_merkle_proof
              ,tx_tree_root ,signature_id ,proposal_block_id ,created_at
              FROM tx_merkle_proofs WHERE id = $1 `
	)

	var tmp models.TxMerkleProofs
	err := errPgx.Err(p.queryRow(p.ctx, q, id).
		Scan(
			&tmp.ID,
			&tmp.SenderPublicKey,
			&tmp.TxHash,
			&tmp.TxTreeIndex,
			&tmp.TxMerkleProof,
			&tmp.TxTreeRoot,
			&tmp.SignatureID,
			&tmp.ProposalBlockID,
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
		q = ` SELECT
              id ,sender_public_key ,tx_hash ,tx_tree_index ,tx_merkle_proof
              ,tx_tree_root ,signature_id ,proposal_block_id ,created_at             
              FROM tx_merkle_proofs WHERE tx_hash = $1 `
	)

	var tmp models.TxMerkleProofs
	err := errPgx.Err(p.queryRow(p.ctx, q, txHash).
		Scan(
			&tmp.ID,
			&tmp.SenderPublicKey,
			&tmp.TxHash,
			&tmp.TxTreeIndex,
			&tmp.TxMerkleProof,
			&tmp.TxTreeRoot,
			&tmp.SignatureID,
			&tmp.ProposalBlockID,
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
		TxTreeIndex:     tmp.TxTreeIndex,
		TxMerkleProof:   tmp.TxMerkleProof,
		TxTreeRoot:      tmp.TxTreeRoot,
		ProposalBlockID: tmp.ProposalBlockID,
		SignatureID:     tmp.SignatureID.String,
		CreatedAt:       tmp.CreatedAt,
	}

	return m
}
