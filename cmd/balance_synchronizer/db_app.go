package balance_synchronizer

import (
	"context"
	"encoding/json"
	"intmax2-node/internal/intmax_block_content"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_db_app.go -package=balance_synchronizer -source=db_app.go

type SQLDriverApp interface {
	GenericCommandsApp
	ServiceCommands
	CtrlProcessingJobs
	Blocks
	Signatures
	TxMerkleProofs
	EventBlockNumbers
	EventBlockNumbersForValidityProver
	CtrlEventBlockNumbersJobs
	EventBlockNumbersErrors
	BlockSenders
	BlockAccounts
	BlockParticipants
	Deposits
	BlockContents
	Withdrawals
	EthereumCounterparties
}

type GenericCommandsApp interface {
	Exec(ctx context.Context, input interface{}, executor func(d interface{}, input interface{}) error) (err error)
}

type ServiceCommands interface {
	Check(ctx context.Context) health.Health
}

type CtrlProcessingJobs interface {
	CreateCtrlProcessingJobs(name string, options json.RawMessage) error
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

type BlockSenders interface {
	CreateBlockSenders(
		address, publicKey string,
	) (*mDBApp.BlockSender, error)
	BlockSenderByAddress(address string) (*mDBApp.BlockSender, error)
}

type BlockAccounts interface {
	CreateBlockAccount(senderID string) (*mDBApp.BlockAccount, error)
	BlockAccountBySenderID(senderID string) (*mDBApp.BlockAccount, error)
	ResetSequenceByBlockAccounts() error
	DelAllBlockAccounts() error
}

type BlockParticipants interface {
	CreateBlockParticipant(
		blockNumber uint32,
		senderId string,
	) (*mDBApp.BlockParticipant, error)
}

type Deposits interface {
	CreateDeposit(depositLeaf intMaxTree.DepositLeaf, depositID uint32, sender string) (*mDBApp.Deposit, error)
	UpdateDepositIndexByDepositHash(depositHash common.Hash, depositIndex uint32) error
	UpdateSenderByDepositID(depositID uint32, sender string) error
	Deposit(ID string) (*mDBApp.Deposit, error)
	DepositByDepositID(depositID uint32) (*mDBApp.Deposit, error)
	DepositByDepositHash(depositHash common.Hash) (*mDBApp.Deposit, error)
	DepositsListByDepositHash(depositHash ...common.Hash) ([]*mDBApp.Deposit, error)
	ScanDeposits() ([]*mDBApp.Deposit, error)
	FetchNextDepositIndex() (uint32, error)
}

type BlockContents interface {
	CreateBlockContent(
		postedBlock *intmax_block_content.PostedBlock,
		blockContent *intMaxTypes.BlockContent,
		l2BlockNumber *uint256.Int,
		l2BlockHash common.Hash,
	) (*mDBApp.BlockContentWithProof, error)
	BlockContentUpdDepositLeavesCounterByBlockNumber(
		blockNumber, depositLeavesCounter uint32,
	) error
	BlockContentByBlockNumber(blockNumber uint32) (*mDBApp.BlockContentWithProof, error)
	BlockContentByTxRoot(txRoot common.Hash) (*mDBApp.BlockContentWithProof, error)
	BlockContentListByTxRoot(txRoot ...common.Hash) ([]*mDBApp.BlockContentWithProof, error)
	ScanBlockHashAndSenders() (blockHashAndSendersMap map[uint32]mDBApp.BlockHashAndSenders, lastBlockNumber uint32, err error)
	CreateValidityProof(blockHash common.Hash, validityProof []byte) (*mDBApp.BlockProof, error)
	LastBlockValidityProof() (*mDBApp.BlockContentWithProof, error)
	LastBlockNumberGeneratedValidityProof() (uint32, error)
	LastPostedBlockNumber() (uint32, error)
}

type Withdrawals interface {
	CreateWithdrawal(
		id string,
		transferData *mDBApp.TransferData,
		transferMerkleProof *mDBApp.TransferMerkleProof,
		transaction *mDBApp.Transaction,
		txMerkleProof *mDBApp.TxMerkleProof,
		transferHash string,
		blockNumber int64,
		blockHash string,
		enoughBalanceProof *mDBApp.EnoughBalanceProof,
	) (*mDBApp.Withdrawal, error)
	UpdateWithdrawalsStatus(ids []string, status mDBApp.WithdrawalStatus) error
	WithdrawalByID(id string) (*mDBApp.Withdrawal, error)
	WithdrawalsByHashes(transferHashes []string) (*[]mDBApp.Withdrawal, error)
	WithdrawalsByStatus(status mDBApp.WithdrawalStatus, limit *int) (*[]mDBApp.Withdrawal, error)
}

type EthereumCounterparties interface {
	CreateEthereumCounterparty(
		address string,
	) (*mDBApp.EthereumCounterparty, error)
	EthereumCounterpartyByAddress(address string) (*mDBApp.EthereumCounterparty, error)
}
