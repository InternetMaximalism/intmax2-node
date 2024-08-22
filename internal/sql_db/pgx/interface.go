package pgx

import (
	"context"
	"encoding/json"
	mFL "intmax2-node/internal/sql_filter/models"
	intMaxTypes "intmax2-node/internal/types"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/holiman/uint256"
)

type PGX interface {
	GenericCommands
	ServiceCommands
	Blocks
	Tokens
	Signatures
	TxMerkleProofs
	EventBlockNumbers
	Balances
	Withdrawals
	BackupTransfers
	BackupTransactions
	BackupDeposits
	CtrlEventBlockNumbersJobs
	EventBlockNumbersErrors
	Senders
	Accounts
	BackupBalances
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

type Tokens interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*mDBApp.Token, error)
	TokenByIndex(tokenIndex string) (*mDBApp.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*mDBApp.Token, error)
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

type Balances interface {
	CreateBalance(userAddress, tokenAddress, balance string) (*mDBApp.Balance, error)
	UpdateBalanceByID(balanceID, balance string) (*mDBApp.Balance, error)
	BalanceByID(id string) (*mDBApp.Balance, error)
	BalanceByUserAndTokenIndex(userAddress, tokenIndex string) (*mDBApp.Balance, error)
	BalanceByUserAndTokenInfo(userAddress, tokenAddress, tokenID string) (*mDBApp.Balance, error)
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

type BackupTransfers interface {
	CreateBackupTransfer(
		recipient, encryptedTransferHash, encryptedTransfer string,
		blockNumber int64,
	) (*mDBApp.BackupTransfer, error)
	GetBackupTransfer(condition string, value string) (*mDBApp.BackupTransfer, error)
	GetBackupTransfers(condition string, value interface{}) ([]*mDBApp.BackupTransfer, error)
}

type BackupTransactions interface {
	CreateBackupTransaction(
		sender, encryptedTxHash, encryptedTx, signature string,
		blockNumber int64,
	) (*mDBApp.BackupTransaction, error)
	GetBackupTransaction(condition string, value string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactionBySenderAndTxDoubleHash(sender, txDoubleHash string) (*mDBApp.BackupTransaction, error)
	GetBackupTransactions(condition string, value interface{}) ([]*mDBApp.BackupTransaction, error)
	GetBackupTransactionsBySender(
		sender string,
		pagination mDBApp.PaginationOfListOfBackupTransactionsInput,
		sorting mFL.Sorting, orderBy mFL.OrderBy,
		filters mFL.FiltersList,
	) (
		paginator *mDBApp.PaginationOfListOfBackupTransactions,
		listDBApp mDBApp.ListOfBackupTransaction,
		err error,
	)
}

type BackupDeposits interface {
	CreateBackupDeposit(
		recipient, depositHash, encryptedDeposit string,
		blockNumber int64,
	) (*mDBApp.BackupDeposit, error)
	GetBackupDeposit(conditions []string, values []interface{}) (*mDBApp.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*mDBApp.BackupDeposit, error)
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

type BackupBalances interface {
	CreateBackupBalance(
		user, encryptedBalanceProof, encryptedBalanceData, signature string,
		encryptedTxs, encryptedTransfers, encryptedDeposits []string,
		blockNumber int64,
	) (*mDBApp.BackupBalance, error)
	GetBackupBalance(conditions []string, values []interface{}) (*mDBApp.BackupBalance, error)
	GetBackupBalances(condition string, value interface{}) ([]*mDBApp.BackupBalance, error)
}
