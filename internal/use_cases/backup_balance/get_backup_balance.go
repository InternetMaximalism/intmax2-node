package backup_balance

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_balance.go -package=mocks -source=get_backup_balance.go

type UCGetBackupBalance struct {
	EncryptedBalanceProof string   `json:"encryptedBalanceProof"`
	EncryptedBalanceData  string   `json:"encryptedBalanceData"`
	EncryptedTxs          []string `json:"encryptedTxs"`
	EncryptedTransfers    []string `json:"encryptedTransfers"`
	EncryptedDeposits     []string `json:"encryptedDeposits"`
}

type UCGetBackupBalanceInput struct {
	User        string               `json:"user"`
	DecodeUser  *intMaxAcc.PublicKey `json:"-"`
	BlockNumber uint32               `json:"blockNumber"`
}

// UseCaseGetBackupBalance describes GetBackupBalance contract.
type UseCaseGetBackupBalance interface {
	Do(ctx context.Context, input *UCGetBackupBalanceInput) (*UCGetBackupBalance, error)
}
