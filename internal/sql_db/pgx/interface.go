package pgx

import (
	"context"
	"encoding/json"
	postWithdrwalRequest "intmax2-node/internal/use_cases/post_withdrawal_request"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"

	"github.com/dimiro1/health"
	"github.com/holiman/uint256"
)

type PGX interface {
	GenericCommands
	ServiceCommands
	Tokens
	Signatures
	Transactions
	TxMerkleProofs
	EventBlockNumbers
	Balances
	Withdrawals
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
	) (*mDBApp.Token, error)
	TokenByIndex(tokenIndex string) (*mDBApp.Token, error)
	TokenByTokenInfo(tokenAddress, tokenID string) (*mDBApp.Token, error)
}

type Signatures interface {
	CreateSignature(signature string) (*mDBApp.Signature, error)
	SignatureByID(txID string) (*mDBApp.Signature, error)
}

type Transactions interface {
	CreateTransaction(
		senderPublicKey, txHash, signatureID string,
	) (*mDBApp.Transactions, error)
	TransactionByID(txID string) (*mDBApp.Transactions, error)
}

type TxMerkleProofs interface {
	CreateTxMerkleProofs(
		senderPublicKey, txHash, txID string,
		txTreeIndex *uint256.Int,
		txMerkleProof json.RawMessage,
	) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByID(id string) (*mDBApp.TxMerkleProofs, error)
	TxMerkleProofsByTxHash(txHash string) (*mDBApp.TxMerkleProofs, error)
}

type EventBlockNumbers interface {
	UpsertEventBlockNumber(eventName string, blockNumber int64) (*mDBApp.EventBlockNumber, error)
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
	CreateWithdrawal(id string, input postWithdrwalRequest.UCPostWithdrawalRequestInput) (*mDBApp.Withdrawal, error)
	WithdrawalByID(id string) (*mDBApp.Withdrawal, error)
	WithdrawalsByStatus(status mDBApp.WithdrawalStatus) (*[]mDBApp.Withdrawal, error)
}
