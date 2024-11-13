package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

func (p *pgx) CreateBlockParticipant(
	blockNumber uint32,
	senderId string,
) (*mDBApp.BlockParticipant, error) {
	const (
		q = `INSERT INTO block_participants (
             block_number ,sender_id
             ) VALUES ($1, $2)
             ON CONFLICT (block_number ,sender_id) DO NOTHING `
	)

	_, err := p.exec(p.ctx, q, blockNumber, senderId)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.BlockParticipant
	bDBApp, err = p.BlockParticipantByBlockNumberAndSenderID(blockNumber, senderId)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) BlockParticipantByBlockNumberAndSenderID(
	blockNumber uint32,
	senderId string,
) (*mDBApp.BlockParticipant, error) {
	const (
		q = `SELECT id, block_number, sender_id, created_at
			 FROM block_participants
			 WHERE block_number = $1 AND sender_id = $2 `
	)

	var tmp models.BlockParticipant
	err := errPgx.Err(p.queryRow(p.ctx, q, blockNumber, senderId).
		Scan(
			&tmp.ID,
			&tmp.BlockNumber,
			&tmp.SenderId,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.blockParticipantToDBApp(&tmp)

	return bDBApp, nil
}

func (p *pgx) blockParticipantToDBApp(tmp *models.BlockParticipant) *mDBApp.BlockParticipant {
	m := mDBApp.BlockParticipant{
		ID:          tmp.ID,
		BlockNumber: uint32(tmp.BlockNumber),
		SenderId:    tmp.SenderId,
		CreatedAt:   tmp.CreatedAt,
	}

	return &m
}
