package server

import (
	"context"
	"encoding/json"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app.go -package=server -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	Blocks
	Signatures
	TxMerkleProofs
	EventBlockNumbers
	EventBlockNumbersForValidityProver
	CtrlEventBlockNumbersJobs
	EventBlockNumbersErrors
	Senders
	Accounts
	Deposits
	BlockContents
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type Blocks interface {
	CreateBlock(
		builderPublicKey, txRoot, aggregatedSignature, aggregatedPublicKey string, senders []intMaxTypes.ColumnSender,
		senderType uint,
		options []byte,
	) (*mDBApp.Block, error)
	Block(proposalBlockID string) (*mDBApp.Block, error)
	BlockByTxRoot(txRoot string) (*mDBApp.Block, error)
	UpdateBlockStatus(proposalBlockID string, blockHash string, blockNumber uint32) error
	GetUnprocessedBlocks() ([]*mDBApp.Block, error)
}

type Signatures interface {
	CreateSignature(signature, proposalBlockID string) (*mDBApp.Signature, error)
	SignatureByID(signatureID string) (*mDBApp.Signature, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, signatureID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
		txTreeRoot string,
		proposalBlockID string,
	) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*mDBApp.TxMerkleProofs, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*mDBApp.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*mDBApp.EventBlockNumber, error)
}

type EventBlockNumbersForValidityProver interface {
	UpsertEventBlockNumberForValidityProver(eventName string, blockNumber uint64) (*mDBApp.EventBlockNumberForValidityProver, error)
	EventBlockNumberByEventNameForValidityProver(eventName string) (*mDBApp.EventBlockNumberForValidityProver, error)
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

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32) (*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	Deposit(ID string) (*mDBApp.Deposit, error)
	DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchLastDepositIndex() (uint32, error)
}

type BlockContents interface {
	// CreateBlockContent(
	// 	blockNumber uint32,
	// 	blockHash, prevBlockHash, depositRoot, txRoot, aggregatedSignature, aggregatedPublicKey, messagePoint string,
	// 	isRegistrationBlock bool,
	// 	senders []intMaxTypes.ColumnSender,
	// ) (*mDBApp.BlockContent, error)
	CreateBlockContent(
		postedBlock *block_post_service.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
	) (*mDBApp.BlockContent, error)
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContent, error)
	// SetValidityWitness(blockNumber uint32, witness *block_validity_prover.ValidityWitness) error
	// LastValidityWitness() (*block_validity_prover.ValidityWitness, error)
	// SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	// LastSeenBlockPostedEventBlockNumber() (blockNumber uint64, err error)
}
