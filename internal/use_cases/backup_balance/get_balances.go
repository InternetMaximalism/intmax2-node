package backup_balance

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_balances.go -package=mocks -source=get_balances.go

type BackupDeposit struct {
	Recipient        string    `json:"recipient"`
	EncryptedDeposit string    `json:"encrypted_deposit"`
	BlockNumber      uint64    `json:"block_number"`
	CreatedAt        time.Time `json:"created_at"`
}

type BackupTransfer struct {
	EncryptedTransfer                string    `json:"encrypted_transfer"`
	Recipient                        string    `json:"recipient"`
	BlockNumber                      uint64    `json:"block_number"`
	SenderLastBalanceProofBody       string    `json:"senderLastBalanceProofBody,omitempty"`
	SenderBalanceTransitionProofBody string    `json:"senderBalanceTransitionProofBody,omitempty"`
	CreatedAt                        time.Time `json:"created_at"`
}

type BackupTransaction struct {
	Sender      string    `json:"sender"`
	EncryptedTx string    `json:"encrypted_tx"`
	BlockNumber uint64    `json:"block_number"`
	CreatedAt   time.Time `json:"created_at"`
}

type UCGetBalances struct {
	Deposits     []*BackupDeposit     `json:"deposits"`
	Transfers    []*BackupTransfer    `json:"transfers"`
	Transactions []*BackupTransaction `json:"transactions"`
}

type UCGetBalancesInput struct {
	Address string `json:"address"`
}

// UseCaseGetBalances describes GetBalances contract.
type UseCaseGetBalances interface {
	Do(ctx context.Context, input *UCGetBalancesInput) (*UCGetBalances, error)
}
