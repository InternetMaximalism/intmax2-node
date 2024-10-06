package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBlockParticipant(
	blockNumber uint32,
	senderId string,
) (*mDBApp.BlockContainedSender, error) {
	s := models.BlockContainedSender{
		BlockContainedSenderID: uuid.New().String(),
		BlockNumber:            int64(blockNumber),
		SenderId:               senderId,
		CreatedAt:              time.Now().UTC(),
	}

	const (
		q = `INSERT INTO block_participants (
             id ,block_number ,sender_id ,created_at
             ) VALUES ($1, $2, $3, $4) `
	)

	_, err := p.exec(p.ctx, q,
		s.BlockContainedSenderID, s.BlockNumber, s.SenderId, s.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.BlockContainedSender
	bDBApp, err = p.BlockParticipant(s.BlockContainedSenderID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) CreateBlockParticipants(
	blockNumber uint32,
	senderPublicKeys []string,
) ([]*mDBApp.BlockContainedSender, error) {
	var senders []*mDBApp.BlockContainedSender
	now := time.Now().UTC()

	const (
		q = `WITH inserted AS (
				INSERT INTO block_participants (id, block_number, sender_id, created_at)
				SELECT $1, $2, senders.id, $4
				FROM senders
				WHERE senders.public_key = $3
				RETURNING id
			 )
			 SELECT id FROM inserted`
	)

	for _, senderPublicKey := range senderPublicKeys {
		id := uuid.New().String()
		_, err := p.exec(p.ctx, q,
			id, int64(blockNumber), senderPublicKey, now)
		if err != nil {
			return nil, errPgx.Err(err)
		}

		sender, err := p.BlockParticipant(id)
		if err != nil {
			return nil, err
		}
		senders = append(senders, sender)
	}

	return senders, nil
}

func (p *pgx) BlockParticipant(blockContentID string) (*mDBApp.BlockContainedSender, error) {
	const (
		q = `SELECT id, block_number, sender_id, created_at
			 FROM block_participants
			 WHERE id = $1`
	)

	var tmp models.BlockContainedSender
	err := errPgx.Err(p.queryRow(p.ctx, q, blockContentID).
		Scan(
			&tmp.BlockContainedSenderID,
			&tmp.BlockNumber,
			&tmp.SenderId,
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
		BlockNumber:            uint32(tmp.BlockNumber),
		SenderId:               tmp.SenderId,
		CreatedAt:              tmp.CreatedAt,
	}

	return &m
}
