package deposit_status_by_hash

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=deposit_status_by_hash_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	Deposits
	// DepositTreeBuilder
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type Deposits interface {
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
}

// type DepositTreeBuilder interface {
// 	LastDepositTreeRoot() (common.Hash, error)
// 	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, error)
// 	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *block_validity_prover.DepositLeafWithId, depositIndex *uint32, err error)
// }
