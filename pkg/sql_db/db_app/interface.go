package db_app

import (
	"context"
	"encoding/json"
	"intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/holiman/uint256"
)

type SQLDb interface {
	GenericCommands
	ServiceCommands
	Tokens
	TxMerkleProofs
}

type GenericCommands interface {
	Begin(ctx context.Context) (interface{}, error)
	Rollback()
	Commit() error
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Migrator(ctx context.Context, command string) (step int, err error)
	Check(ctx context.Context) health.Health
}

type Tokens interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*models.Token, error)
	TokenByIndex(tokenIndex string) (*models.Token, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
	) (*models.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*models.TxMerkleProofs, error)
}
