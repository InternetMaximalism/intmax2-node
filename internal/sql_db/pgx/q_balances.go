package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"time"

	"github.com/google/uuid"
)

func (p *pgx) CreateBalance(userAddress, tokenIndex, balance string) (*mDBApp.Balance, error) {
	s := models.Balance{
		ID:          uuid.New().String(),
		UserAddress: userAddress,
		TokenIndex:  tokenIndex,
		Balance:     balance,
		CreatedAt:   time.Now().UTC(),
	}

	const (
		q = ` INSERT INTO balances
              (id, user, token_index, balance, created_at)
              VALUES ($1, $2, $3, $4, $5) `
	)

	_, err := p.exec(p.ctx, q,
		s.ID, s.UserAddress, s.TokenIndex, s.Balance, s.CreatedAt)
	if err != nil {
		return nil, errPgx.Err(err)
	}

	var bDBApp *mDBApp.Balance
	bDBApp, err = p.BalanceByID(s.ID)
	if err != nil {
		return nil, err
	}

	return bDBApp, nil
}

func (p *pgx) BalanceByID(balanceID string) (*mDBApp.Balance, error) {
	const (
		q = ` SELECT id, user, token_index, balance, created_at
              FROM balances WHERE id = $1 `
	)

	var tmp models.Balance
	err := errPgx.Err(p.queryRow(p.ctx, q, balanceID).
		Scan(
			&tmp.ID,
			&tmp.UserAddress,
			&tmp.TokenIndex,
			&tmp.Balance,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.bToDBApp(&tmp)

	return &bDBApp, nil
}

func (p *pgx) BalanceByUserAndTokenIndex(userAddress, tokenIndex string) (*mDBApp.Balance, error) {
	const (
		q = ` SELECT id, user, token_index, balance, created_at
              FROM balances WHERE user_address = $1 AND token_index = $2 `
	)

	var tmp models.Balance
	err := errPgx.Err(p.queryRow(p.ctx, q, userAddress, tokenIndex).
		Scan(
			&tmp.ID,
			&tmp.UserAddress,
			&tmp.TokenIndex,
			&tmp.Balance,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.bToDBApp(&tmp)

	return &bDBApp, nil
}

func (p *pgx) BalanceByUserAndTokenInfo(userAddress, tokenAddress string, tokenID string) (*mDBApp.Balance, error) {
	const (
		q = ` SELECT b.id, b.user_address, b.token_index, b.balance, b.created_at
			  FROM balances b
			  JOIN tokens t ON b.token_index = t.token_index
			  WHERE b.user_address = $1
			  AND t.token_address = $2
			  AND t.token_id = $3 `
	)

	var tmp models.Balance
	err := errPgx.Err(p.queryRow(p.ctx, q, userAddress, tokenAddress, tokenID).
		Scan(
			&tmp.ID,
			&tmp.UserAddress,
			&tmp.TokenIndex,
			&tmp.Balance,
			&tmp.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	bDBApp := p.bToDBApp(&tmp)

	return &bDBApp, nil
}

func (p *pgx) bToDBApp(b *models.Balance) mDBApp.Balance {
	m := mDBApp.Balance{
		ID:          b.ID,
		UserAddress: b.UserAddress,
		TokenIndex:  b.TokenIndex,
		Balance:     b.Balance,
		CreatedAt:   b.CreatedAt,
	}

	return m
}
