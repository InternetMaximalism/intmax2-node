package block_validity_prover_server

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_server_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommandsApp
	BlockContents
	RelationshipL2BatchIndexAndBlockContent
	L2BatchIndex
	Senders
	Accounts
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommandsApp interface {
	Check(ctx context.Context) health.Health
}

type BlockContents interface {
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockHash(blockHash string) (*mDBApp.BlockContentWithProof, error)
}

type RelationshipL2BatchIndexAndBlockContent interface {
	RelationshipL2BatchIndexAndBlockContentsByBlockContentID(
		blockContentID string,
	) (*mDBApp.RelationshipL2BatchIndexBlockContents, error)
}

type L2BatchIndex interface {
	L2BatchIndex(batchIndex *uint256.Int) (*mDBApp.L2BatchIndex, error)
}

type Senders interface {
	SenderByAddress(address string) (*mDBApp.Sender, error)
}

type Accounts interface {
	AccountBySenderID(senderID string) (*mDBApp.Account, error)
}
