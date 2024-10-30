package server

import (
	"context"
	block_post_service "intmax2-node/internal/block_post_service"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	uint256 "github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app.go -package=server -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	Blocks
	Deposits
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type Blocks interface {
	BlockByTxRoot(txRoot string) (*mDBApp.Block, error)
}

type Deposits interface {
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchNextDepositIndex() (uint32, error)
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	ScanBlockHashAndSenders() (blockHashAndSendersMap map[uint32]mDBApp.BlockHashAndSenders, lastBlockNumber uint32, err error)
	CreateValidityProof(blockHash common.Hash, validityProof []byte) (*mDBApp.BlockProof, error)
	LastBlockValidityProof() (*mDBApp.BlockContentWithProof, error)
	LastBlockNumberGeneratedValidityProof() (uint32, error)
	LastPostedBlockNumber() (uint32, error)
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string) error
	CtrlProcessingJobs(name string) (*mDBApp.CtrlProcessingJobs, error)
}

type GasPriceOracleApp interface {
	CreateGasPriceOracle(name string, value *uint256.Int) error
	GasPriceOracle(name string) (*mDBApp.GasPriceOracle, error)
}
