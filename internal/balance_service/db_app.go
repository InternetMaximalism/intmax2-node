package balance_service

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=balance_service_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	TokensApp
	BalanceApp
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type TokensApp interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*mDBApp.Token, error)
	TokenByIndex(tokenIndex string) (*mDBApp.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*mDBApp.Token, error)
}

type BalanceApp interface {
	CreateBalance(userAddress, tokenAddress, balance string) (*mDBApp.Balance, error)
	UpdateBalanceByID(balanceID, balance string) (*mDBApp.Balance, error)
	BalanceByUserAndTokenIndex(userAddress, tokenIndex string) (*mDBApp.Balance, error)
	BalanceByUserAndTokenInfo(userAddress, tokenAddress, tokenID string) (*mDBApp.Balance, error)
}
