package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"
)

func (p *pgx) CreateBlockSenders(
	address, publicKey string,
) (*mDBApp.BlockSender, error) {
	const (
		q = ` INSERT INTO block_senders
              (address ,public_key ,created_at)
              VALUES ($1 ,$2 ,$3) `
	)

	addressWithoutPrefix := strings.TrimPrefix(address, "0x")
	publicKeyWithoutPrefix := strings.TrimPrefix(publicKey, "0x")

	_, err := p.exec(p.ctx, q, addressWithoutPrefix, publicKeyWithoutPrefix, time.Now().UTC())
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var senderDBApp *mDBApp.BlockSender
	senderDBApp, err = p.BlockSenderByAddress(addressWithoutPrefix)
	if err != nil {
		return nil, err
	}

	return senderDBApp, nil
}

func (p *pgx) BlockSenderByID(id string) (*mDBApp.BlockSender, error) {
	const (
		q = ` SELECT id ,address ,public_key ,created_at
              FROM block_senders
              WHERE id = $1 `
	)

	var sender models.BlockSender
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

	senderDBApp := p.blockSenderToDBApp(&sender)

	return &senderDBApp, nil
}

func (p *pgx) BlockSenderByAddress(address string) (*mDBApp.BlockSender, error) {
	const (
		q = ` SELECT id ,address ,public_key ,created_at
              FROM block_senders
              WHERE address = $1 `
	)

	var sender models.BlockSender
	err := errPgx.Err(p.queryRow(p.ctx, q, strings.TrimPrefix(address, "0x")).
		Scan(
			&sender.ID,
			&sender.Address,
			&sender.PublicKey,
			&sender.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	senderDBApp := p.blockSenderToDBApp(&sender)

	return &senderDBApp, nil
}

func (p *pgx) BlockSenderByPublicKey(publicKey string) (*mDBApp.BlockSender, error) {
	const (
		q = ` SELECT id ,address ,public_key ,created_at
              FROM block_senders
              WHERE public_key = $1 `
	)

	var sender models.BlockSender
	err := errPgx.Err(p.queryRow(p.ctx, q, strings.TrimPrefix(publicKey, "0x")).
		Scan(
			&sender.ID,
			&sender.Address,
			&sender.PublicKey,
			&sender.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	senderDBApp := p.blockSenderToDBApp(&sender)

	return &senderDBApp, nil
}

func (p *pgx) blockSenderToDBApp(sender *models.BlockSender) mDBApp.BlockSender {
	return mDBApp.BlockSender{
		ID:        sender.ID,
		Address:   sender.Address,
		PublicKey: sender.PublicKey,
		CreatedAt: sender.CreatedAt,
	}
}
