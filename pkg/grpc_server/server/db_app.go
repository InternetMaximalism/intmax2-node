package server

import (
	"context"
	"encoding/json"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=server_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	TxMerkleProofs
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, txID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
	) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*mDBApp.TxMerkleProofs, error)
}
