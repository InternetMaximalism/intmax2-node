package pgx

import (
	intMaxAcc "intmax2-node/internal/accounts"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

func (p *pgx) CreateBlockAccount(senderID string) (*mDBApp.BlockAccount, error) {
	const (
		q = ` INSERT INTO block_accounts (sender_id) VALUES ($1) `
	)

	_, err := p.exec(p.ctx, q, senderID)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var accountDBApp *mDBApp.BlockAccount
	accountDBApp, err = p.BlockAccountBySenderID(senderID)
	if err != nil {
		return nil, err
	}

	return accountDBApp, nil
}

func (p *pgx) BlockAccountBySenderID(senderID string) (*mDBApp.BlockAccount, error) {
	const (
		q = ` SELECT id ,account_id ,sender_id ,created_at
              FROM block_accounts WHERE sender_id = $1 `
	)

	var account models.BlockAccount
	err := errPgx.Err(p.queryRow(p.ctx, q, senderID).
		Scan(
			&account.ID,
			&account.AccountID,
			&account.SenderID,
			&account.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	accountDBApp := p.blockAccountToDBApp(&account)

	return &accountDBApp, nil
}

func (p *pgx) BlockAccountBySender(publicKey *intMaxAcc.PublicKey) (*mDBApp.BlockAccount, error) {
	address := publicKey.ToAddress().String()
	const (
		q = ` SELECT a.id, a.account_id, a.sender_id, a.created_at
              FROM block_accounts a
			  JOIN block_senders s ON a.sender_id = s.id
			  WHERE s.address = $1 `
	)

	var account models.BlockAccount
	err := errPgx.Err(p.queryRow(p.ctx, q, address).
		Scan(
			&account.ID,
			&account.AccountID,
			&account.SenderID,
			&account.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	accountDBApp := p.blockAccountToDBApp(&account)

	return &accountDBApp, nil
}

func (p *pgx) BlockAccountByAccountID(accountID *uint256.Int) (*mDBApp.BlockAccount, error) {
	const (
		q = ` SELECT id ,account_id ,sender_id ,created_at
              FROM block_accounts WHERE account_id = $1 `
	)

	cID, _ := accountID.Value()

	var account models.BlockAccount
	err := errPgx.Err(p.queryRow(p.ctx, q, cID).
		Scan(
			&account.ID,
			&account.AccountID,
			&account.SenderID,
			&account.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	accountDBApp := p.blockAccountToDBApp(&account)

	return &accountDBApp, nil
}

func (p *pgx) ResetSequenceByBlockAccounts() error {
	const (
		q = ` ALTER SEQUENCE block_accounts_account_id_seq RESTART WITH 2 `
	)

	_, err := p.exec(p.ctx, q)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) DelAllBlockAccounts() error {
	const (
		q = ` DELETE FROM block_accounts WHERE 1=1 `
	)

	_, err := p.exec(p.ctx, q)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) blockAccountToDBApp(account *models.BlockAccount) mDBApp.BlockAccount {
	return mDBApp.BlockAccount{
		ID:        account.ID,
		AccountID: &account.AccountID,
		SenderID:  account.SenderID,
		CreatedAt: account.CreatedAt,
	}
}
