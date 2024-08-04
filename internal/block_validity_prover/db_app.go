package block_validity_prover

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app_test.go -package=block_validity_prover_test -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	Blocks
	EventBlockNumbers
	CtrlEventBlockNumbersJobs
	EventBlockNumbersErrors
	Senders
	Accounts
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type Blocks interface {
	UpdateBlockStatus(proposalBlockID string, status int64) error
	GetUnprocessedBlocks() ([]*mDBApp.Block, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*mDBApp.EventBlockNumber, error)
}

type CtrlEventBlockNumbersJobs interface {
	CreateCtrlEventBlockNumbersJobs(eventName string) error
	CtrlEventBlockNumbersJobs(eventName string) (*mDBApp.CtrlEventBlockNumbersJobs, error)
}

type EventBlockNumbersErrors interface {
	UpsertEventBlockNumbersErrors(
		eventName string,
		blockNumber *uint256.Int,
		options []byte,
		updErr error,
	) error
	EventBlockNumbersErrors(
		eventName string,
		blockNumber *uint256.Int,
	) (*mDBApp.EventBlockNumbersErrors, error)
}

type Senders interface {
	CreateSenders(
		address, publicKey string,
	) (*mDBApp.Sender, error)
	SenderByID(id string) (*mDBApp.Sender, error)
	SenderByAddress(address string) (*mDBApp.Sender, error)
	SenderByPublicKey(publicKey string) (*mDBApp.Sender, error)
}

type Accounts interface {
	CreateAccount(senderID string) (*mDBApp.Account, error)
	AccountBySenderID(senderID string) (*mDBApp.Account, error)
	AccountByAccountID(accountID *uint256.Int) (*mDBApp.Account, error)
	ResetSequenceByAccounts() error
	DelAllAccounts() error
}
