package db_app

import (
	"context"
	"encoding/json"
	intMaxTypes "intmax2-node/internal/types"
	backupBalance "intmax2-node/internal/use_cases/backup_balance"
	backupDeposit "intmax2-node/internal/use_cases/backup_deposit"
	backupTransaction "intmax2-node/internal/use_cases/backup_transaction"
	backupTransfer "intmax2-node/internal/use_cases/backup_transfer"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	"intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/holiman/uint256"
)

type SQLDb interface {
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
	) (*models.Block, error)
	Block(proposalBlockID string) (*models.Block, error)
	BlockByTxRoot(txRoot string) (*models.Block, error)
	UpdateBlockStatus(proposalBlockID string, blockHash string, blockNumber uint32) error
	GetUnprocessedBlocks() ([]*models.Block, error)
}

type Tokens interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*models.Token, error)
	TokenByIndex(tokenIndex string) (*models.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*models.Token, error)
}

type Signatures interface {
	CreateSignature(signature, proposalBlockID string) (*models.Signature, error)
	SignatureByID(signatureID string) (*models.Signature, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, signatureID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
		txTreeRoot string,
		proposalBlockID string,
	) (*models.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*models.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*models.TxMerkleProofs, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber uint64) (*models.EventBlockNumber, error)
	EventBlockNumberByEventName(eventName string) (*models.EventBlockNumber, error)
	EventBlockNumbersByEventNames(eventNames []string) ([]*models.EventBlockNumber, error)
}

type Balances interface {
	CreateBalance(userAddress, tokenAddress, balance string) (*models.Balance, error)
	UpdateBalanceByID(balanceID, balance string) (*models.Balance, error)
	BalanceByID(id string) (*models.Balance, error)
	BalanceByUserAndTokenIndex(userAddress, tokenIndex string) (*models.Balance, error)
	BalanceByUserAndTokenInfo(userAddress, tokenAddress, tokenID string) (*models.Balance, error)
}

type Withdrawals interface {
	CreateWithdrawal(id string, input *postWithdrwalRequest.UCPostWithdrawalRequestInput) (*models.Withdrawal, error)
	UpdateWithdrawalsStatus(ids []string, status models.WithdrawalStatus) error
	WithdrawalByID(id string) (*models.Withdrawal, error)
	WithdrawalsByHashes(transferHashes []string) (*[]models.Withdrawal, error)
	WithdrawalsByStatus(status models.WithdrawalStatus, limit *int) (*[]models.Withdrawal, error)
}

type BackupTransfers interface {
	CreateBackupTransfer(input *backupTransfer.UCPostBackupTransferInput) (*models.BackupTransfer, error)
	GetBackupTransfer(condition string, value string) (*models.BackupTransfer, error)
	GetBackupTransfers(condition string, value interface{}) ([]*models.BackupTransfer, error)
}

type BackupTransactions interface {
	CreateBackupTransaction(input *backupTransaction.UCPostBackupTransactionInput) (*models.BackupTransaction, error)
	GetBackupTransaction(condition string, value string) (*models.BackupTransaction, error)
	GetBackupTransactions(condition string, value interface{}) ([]*models.BackupTransaction, error)
}

type BackupDeposits interface {
	CreateBackupDeposit(input *backupDeposit.UCPostBackupDepositInput) (*models.BackupDeposit, error)
	GetBackupDeposit(conditions []string, values []interface{}) (*models.BackupDeposit, error)
	GetBackupDeposits(condition string, value interface{}) ([]*models.BackupDeposit, error)
}

type CtrlEventBlockNumbersJobs interface {
	CreateCtrlEventBlockNumbersJobs(eventName string) error
	CtrlEventBlockNumbersJobs(eventName string) (*models.CtrlEventBlockNumbersJobs, error)
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
	) (*models.EventBlockNumbersErrors, error)
}

type Senders interface {
	CreateSenders(
		address, publicKey string,
	) (*models.Sender, error)
	SenderByID(id string) (*models.Sender, error)
	SenderByAddress(address string) (*models.Sender, error)
	SenderByPublicKey(publicKey string) (*models.Sender, error)
}

type Accounts interface {
	CreateAccount(senderID string) (*models.Account, error)
	AccountBySenderID(senderID string) (*models.Account, error)
	AccountByAccountID(accountID *uint256.Int) (*models.Account, error)
	ResetSequenceByAccounts() error
	DelAllAccounts() error
}

type BackupBalances interface {
	CreateBackupBalance(input *backupBalance.UCPostBackupBalanceInput) (*models.BackupBalance, error)
	GetBackupBalance(conditions []string, values []interface{}) (*models.BackupBalance, error)
	GetBackupBalances(condition string, value interface{}) ([]*models.BackupBalance, error)
}
