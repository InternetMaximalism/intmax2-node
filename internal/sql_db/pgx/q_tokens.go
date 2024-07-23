package pgx

import (
	errPgx "intmax2-node/internal/sql_db/pgx/errors"
	"intmax2-node/internal/sql_db/pgx/models"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/holiman/uint256"
)

func (p *pgx) CreateToken(
	tokenIndex, tokenAddress string,
	tokenID *uint256.Int,
) (*mDBApp.Token, error) {
	const (
		emptyKey = ""

		qWithTA = ` INSERT INTO tokens
              (id, token_index, token_address, token_id, created_at)
              VALUES ($1, $2, $3, $4, $5) `

		qWithoutTA = ` INSERT INTO tokens
              (id, token_index, token_id, created_at)
              VALUES ($1, $2, $3, $4) `
	)

	id := uuid.New().String()
	tID, _ := tokenID.Value()
	tokenAddress = strings.TrimSpace(tokenAddress)
	createdAt := time.Now().UTC()

	var err error
	if tokenAddress == emptyKey {
		_, err = p.exec(p.ctx, qWithoutTA, id, tokenIndex, tID, createdAt)
		if err != nil {
			return nil, errPgx.Err(err)
		}
	} else {
		_, err = p.exec(p.ctx, qWithTA, id, tokenIndex, tokenAddress, tID, createdAt)
		if err != nil {
			return nil, errPgx.Err(err)
		}
	}

	var tDBApp *mDBApp.Token
	tDBApp, err = p.TokenByIndex(tokenIndex)
	if err != nil {
		return nil, err
	}

	return tDBApp, nil
}

func (p *pgx) TokenByIndex(tokenIndex string) (*mDBApp.Token, error) {
	const (
		q = ` SELECT id, token_index, token_address, token_id, created_at
              FROM tokens WHERE token_index = $1 `
	)

	var t models.Token
	err := errPgx.Err(p.queryRow(p.ctx, q, tokenIndex).
		Scan(
			&t.ID,
			&t.TokenIndex,
			&t.TokenAddress,
			&t.TokenID,
			&t.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	tDBApp := p.tToDBApp(&t)

	return &tDBApp, nil
}

func (p *pgx) TokenByTokenInfo(tokenAddress, tokenID string) (*mDBApp.Token, error) {
	const (
		q = ` SELECT id, token_index, token_address, token_id, created_at
              FROM tokens WHERE token_address = $1 AND token_id = $2`
	)

	var t models.Token
	err := errPgx.Err(p.queryRow(p.ctx, q, tokenAddress, tokenID).
		Scan(
			&t.ID,
			&t.TokenIndex,
			&t.TokenAddress,
			&t.TokenID,
			&t.CreatedAt,
		))
	if err != nil {
		return nil, err
	}

	tDBApp := p.tToDBApp(&t)

	return &tDBApp, nil
}

func (p *pgx) tToDBApp(t *models.Token) mDBApp.Token {
	const emptyKey = ""

	m := mDBApp.Token{
		ID:         t.ID,
		TokenIndex: t.TokenIndex,
		TokenID:    t.TokenID,
		CreatedAt:  t.CreatedAt,
	}

	if ta := strings.TrimSpace(t.TokenAddress.String); ta != emptyKey && t.TokenAddress.Valid {
		m.TokenAddress = ta
	}

	return m
}
