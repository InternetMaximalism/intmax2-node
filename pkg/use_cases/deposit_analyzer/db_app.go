package deposit_analyzer

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=deposit_analyzer_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	TokensApp
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
}
