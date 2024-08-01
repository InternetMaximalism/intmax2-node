package pgx

import (
	"encoding/json"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBlock(
	builderPublicKey, txRoot, aggregatedSignature, aggregatedPublicKey, sendersJSON string,
	senderType uint,
	options []byte,
) (*mDBApp.Block, error) {
	s := models.Block{
		ProposalBlockID:     uuid.New().String(),
		BuilderPublicKey:    builderPublicKey,
		TxRoot:              txRoot,
		AggregatedSignature: aggregatedSignature,
		AggregatedPublicKey: aggregatedPublicKey,
		CreatedAt:           time.Now().UTC(),
		SenderType:          int64(senderType),
		Options:             options,
	}
	if s.Options == nil {
		s.Options = json.RawMessage(`{}`)
	}

	const (
		q = `INSERT INTO blocks (
             proposal_block_id ,builder_public_key ,tx_root
             ,aggregated_signature ,aggregated_public_key
             ,created_at ,sender_type ,options
             ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)  `
	)

	_, err := p.exec(p.ctx, q,
		s.ProposalBlockID, s.BuilderPublicKey, s.TxRoot,
		s.AggregatedSignature, s.AggregatedPublicKey,
		s.CreatedAt, s.SenderType, s.Options)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.Block
	bDBApp, err = p.Block(s.ProposalBlockID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) Block(proposalBlockID string) (*mDBApp.Block, error) {
	const (
		q = `SELECT
             proposal_block_id ,builder_public_key ,tx_root
             ,block_hash ,aggregated_signature ,aggregated_public_key ,status
             ,created_at ,posted_at ,sender_type ,options
             FROM blocks WHERE proposal_block_id = $1`
	)

	var tmp models.Block
	err := errPgx.Err(p.queryRow(p.ctx, q, proposalBlockID).
		Scan(
			&tmp.ProposalBlockID,
			&tmp.BuilderPublicKey,
			&tmp.TxRoot,
			&tmp.BlockHash,
			&tmp.AggregatedSignature,
			&tmp.AggregatedPublicKey,
			&tmp.Status,
			&tmp.CreatedAt,
			&tmp.PostedAt,
			&tmp.SenderType,
			&tmp.Options,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockToDBApp(&tmp)

	return bDBApp, nil
}

// UpdateBlockStatus updates the status of the block with the given proposalBlockID.
func (p *pgx) UpdateBlockStatus(proposalBlockID string, status int64) error {
	const (
		q = `UPDATE blocks SET status = $1 WHERE proposal_block_id = $2`
	)

	_, err := p.exec(p.ctx, q, status, proposalBlockID)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) GetUnprocessedBlocks() ([]*mDBApp.Block, error) {
	const (
		q = `SELECT
			 proposal_block_id ,builder_public_key ,tx_root
			 ,block_hash ,aggregated_signature ,aggregated_public_key ,status
			 ,created_at ,posted_at ,sender_type ,options
			 FROM blocks WHERE status IS NULL`
	)

	rows, err := p.query(p.ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*mDBApp.Block
	for rows.Next() {
		var tmp models.Block
		err = rows.Scan(
			&tmp.ProposalBlockID,
			&tmp.BuilderPublicKey,
			&tmp.TxRoot,
			&tmp.BlockHash,
			&tmp.AggregatedSignature,
			&tmp.AggregatedPublicKey,
			&tmp.Status,
			&tmp.CreatedAt,
			&tmp.PostedAt,
			&tmp.SenderType,
			&tmp.Options,
		)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, p.blockToDBApp(&tmp))
	}

	return blocks, nil
}

func (p *pgx) blockToDBApp(tmp *models.Block) *mDBApp.Block {
	m := mDBApp.Block{
		ProposalBlockID:     tmp.ProposalBlockID,
		BuilderPublicKey:    tmp.BuilderPublicKey,
		TxRoot:              tmp.TxRoot,
		AggregatedSignature: tmp.AggregatedSignature,
		AggregatedPublicKey: tmp.BuilderPublicKey,
		CreatedAt:           tmp.CreatedAt,
		SenderType:          tmp.SenderType,
		Options:             tmp.Options,
	}

	if tmp.BlockHash.Valid && strings.TrimSpace(tmp.BlockHash.String) != "" {
		m.BlockHash = strings.TrimSpace(tmp.BlockHash.String)
	}

	if tmp.Status.Valid {
		m.Status = &tmp.Status.Int64
	}

	if tmp.PostedAt.Valid {
		m.PostedAt = &tmp.PostedAt.Time
	}

	return &m
}
