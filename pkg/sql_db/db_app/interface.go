package db_app

import (
	"context"
	"encoding/json"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	"intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/holiman/uint256"
)

type SQLDb interface {
	GenericCommands
	ServiceCommands
	Tokens
	Withdrawals
	Signatures
	Transactions
	TxMerkleProofs
	EventBlockNumbers
	Balances
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

type Tokens interface {
	CreateToken(
		tokenIndex, tokenAddress string,
		tokenID *uint256.Int,
	) (*models.Token, error)
	TokenByIndex(tokenIndex string) (*models.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*models.Token, error)
}

type Signatures interface {
	CreateSignature(signature string) (*models.Signature, error)
	SignatureByID(txID string) (*models.Signature, error)
}

type Transactions interface {
	CreateTransaction(
		senderPublicKey, txHash, signatureID string,
	) (*models.Transactions, error)
	TransactionByID(txID string) (*models.Transactions, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, txID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
	) (*models.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*models.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*models.TxMerkleProofs, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber int64) (*models.EventBlockNumber, error)
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
