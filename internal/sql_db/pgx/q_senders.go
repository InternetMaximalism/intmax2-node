package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"
)

func (p *pgx) CreateSenders(
	address, publicKey string,
) (*mDBApp.Sender, error) {
	const (
		q = ` INSERT INTO senders
              (address ,public_key ,created_at)
              VALUES ($1 ,$2 ,$3) `
	)

	_, err := p.exec(p.ctx, q, address, publicKey, time.Now().UTC())
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var senderDBApp *mDBApp.Sender
	senderDBApp, err = p.SenderByAddress(address)
	if err != nil {
		return nil, err
	}

	return senderDBApp, nil
}

func (p *pgx) SenderByID(id string) (*mDBApp.Sender, error) {
	const (
		q = ` SELECT id ,address ,public_key ,created_at
              FROM senders
              WHERE id = $1 `
	)

	var sender models.Sender
	err := errPgx.Err(p.queryRow(p.ctx, q, id).
		Scan(
			&sender.ID,
			&sender.Address,
			&sender.PublicKey,
			&sender.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	senderDBApp := p.senderToDBApp(&sender)

	return &senderDBApp, nil
}

func (p *pgx) SenderByAddress(address string) (*mDBApp.Sender, error) {
	const (
		q = ` SELECT id ,address ,public_key ,created_at
              FROM senders
              WHERE address = $1 `
	)

	var sender models.Sender
	err := errPgx.Err(p.queryRow(p.ctx, q, address).
		Scan(
			&sender.ID,
			&sender.Address,
			&sender.PublicKey,
			&sender.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	senderDBApp := p.senderToDBApp(&sender)

	return &senderDBApp, nil
}

func (p *pgx) SenderByPublicKey(publicKey string) (*mDBApp.Sender, error) {
	const (
		q = ` SELECT id ,address ,public_key ,created_at
              FROM senders
              WHERE public_key = $1 `
	)

	var sender models.Sender
	err := errPgx.Err(p.queryRow(p.ctx, q, publicKey).
		Scan(
			&sender.ID,
			&sender.Address,
			&sender.PublicKey,
			&sender.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	senderDBApp := p.senderToDBApp(&sender)

	return &senderDBApp, nil
}

func (p *pgx) senderToDBApp(sender *models.Sender) mDBApp.Sender {
	return mDBApp.Sender{
		ID:        sender.ID,
		Address:   sender.Address,
		PublicKey: sender.PublicKey,
		CreatedAt: sender.CreatedAt,
	}
}
