package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBlockContainedSender(
	blockHash, senderPublicKey string,
	senders []intMaxTypes.ColumnSender,
	senderType uint,
) (*mDBApp.BlockContainedSender, error) {
	s := models.BlockContainedSender{
		BlockContainedSenderID: uuid.New().String(),
		BlockHash:              blockHash,
		Sender:                 senderPublicKey,
		CreatedAt:              time.Now().UTC(),
	}

	const (
		q = `INSERT INTO block_contained_senders (
             id ,block_hash ,sender ,created_at
             ) VALUES ($1, $2, $3, $4) `
	)

	_, err := p.exec(p.ctx, q,
		s.BlockContainedSenderID, s.BlockHash, s.Sender, s.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.BlockContainedSender
	bDBApp, err = p.BlockContainedSender(s.BlockContainedSenderID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) BlockContainedSender(blockContentID string) (*mDBApp.BlockContainedSender, error) {
	const (
		q = `SELECT
             id ,block_hash ,prev_block_hash ,deposit_root
			 ,is_registration_block ,senders ,tx_tree_root ,aggregated_public_key
			 ,aggregated_signature ,message_point ,created_at
             FROM blocks WHERE id = $1`
	)

	var tmp models.BlockContainedSender
	err := errPgx.Err(p.queryRow(p.ctx, q, blockContentID).
		Scan(
			&tmp.BlockContainedSenderID,
			&tmp.BlockHash,
			&tmp.Sender,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockContainedSenderToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) blockContainedSenderToDBApp(tmp *models.BlockContainedSender) *mDBApp.BlockContainedSender {
	m := mDBApp.BlockContainedSender{
		BlockContainedSenderID: tmp.BlockContainedSenderID,
		BlockHash:              tmp.BlockHash,
		Sender:                 tmp.Sender,
		CreatedAt:              tmp.CreatedAt,
	}

	return &m
}
