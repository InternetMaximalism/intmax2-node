package block_builder_storage

import (
	"intmax2-node/internal/block_post_service"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_builder_storage_test -source=db_app.go

type SQLDriverApp interface {
	BlockContents
	Deposits
	EventBlockNumbersForValidityProver
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	ScanBlockHashAndSenders() (
		blockHashAndSendersMap map[uint32]mDBApp.BlockHashAndSenders,
		lastBlockNumber uint32,
		err error,
	)
	LastPostedBlockNumber() (uint32, error)
	LastBlockNumberGeneratedValidityProof() (uint32, error)
	LastBlockValidityProof() (*mDBApp.BlockContentWithProof, error)
	CreateValidityProof(blockHash common.Hash, validityProof []byte) (*mDBApp.BlockProof, error)
}

type Deposits interface {
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchLastDepositIndex() (uint32, error)
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
}

type EventBlockNumbersForValidityProver interface {
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
}
