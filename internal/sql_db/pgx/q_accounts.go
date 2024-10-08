package pgx

import (
	intMaxAcc "intmax2-node/internal/accounts"
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

func (p *pgx) CreateAccount(senderID string) (*mDBApp.Account, error) {
	const (
		q = ` INSERT INTO accounts (sender_id) VALUES ($1) `
	)

	_, err := p.exec(p.ctx, q, senderID)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var accountDBApp *mDBApp.Account
	accountDBApp, err = p.AccountBySenderID(senderID)
	if err != nil {
		return nil, err
	}

	return accountDBApp, nil
}

func (p *pgx) AccountBySenderID(senderID string) (*mDBApp.Account, error) {
	const (
		q = ` SELECT id ,account_id ,sender_id ,created_at
              FROM accounts WHERE sender_id = $1 `
	)

	var account models.Account
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

	accountDBApp := p.accountToDBApp(&account)

	return &accountDBApp, nil
}

func (p *pgx) AccountBySender(publicKey *intMaxAcc.PublicKey) (*mDBApp.Account, error) {
	const (
		q = ` SELECT accounts.id, accounts.account_id, accounts.sender_id, accounts.created_at
              FROM accounts
			  JOIN senders ON accounts.sender_id = senders.id
			  WHERE senders.address = $1 `
	)

	var account models.Account
	err := errPgx.Err(p.queryRow(p.ctx, q, publicKey.ToAddress().String()).
		Scan(
			&account.ID,
			&account.AccountID,
			&account.SenderID,
			&account.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	accountDBApp := p.accountToDBApp(&account)

	return &accountDBApp, nil
}

func (p *pgx) AccountByAccountID(accountID *uint256.Int) (*mDBApp.Account, error) {
	const (
		q = ` SELECT id ,account_id ,sender_id ,created_at
              FROM accounts WHERE account_id = $1 `
	)

	cID, _ := accountID.Value()

	var account models.Account
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

	accountDBApp := p.accountToDBApp(&account)

	return &accountDBApp, nil
}

func (p *pgx) ResetSequenceByAccounts() error {
	const (
		q = ` ALTER SEQUENCE accounts_account_id_seq RESTART WITH 2 `
	)

	_, err := p.exec(p.ctx, q)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) DelAllAccounts() error {
	const (
		q = ` DELETE FROM accounts WHERE 1=1 `
	)

	_, err := p.exec(p.ctx, q)
	if err != nil {
		return errPgx.Err(err)
	}

	return nil
}

func (p *pgx) accountToDBApp(account *models.Account) mDBApp.Account {
	return mDBApp.Account{
		ID:        account.ID,
		AccountID: &account.AccountID,
		SenderID:  account.SenderID,
		CreatedAt: account.CreatedAt,
	}
}
