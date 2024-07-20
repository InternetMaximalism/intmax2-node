package backup_balance

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_balance.go -package=mocks -source=post_backup_balance.go

type UCPostBackupBalance struct {
	Message string `json:"message"`
}

type UCPostBackupBalanceInput struct {
	User                  string               `json:"user"`
	DecodeUser            *intMaxAcc.PublicKey `json:"-"`
	BlockNumber           uint32               `json:"blockNumber"`
	EncryptedBalanceProof string               `json:"encryptedBalanceProof"`
	EncryptedBalanceData  string               `json:"encryptedBalanceData"`
	EncryptedTxs          []string             `json:"encryptedTxs"`
	EncryptedTransfers    []string             `json:"encryptedTransfers"`
	EncryptedDeposits     []string             `json:"encryptedDeposits"`
	Signature             string               `json:"signature"`
}

// UseCasePostBackupBalance describes PostBackupBalance contract.
type UseCasePostBackupBalance interface {
	Do(ctx context.Context, input *UCPostBackupBalanceInput) (*UCPostBackupBalance, error)
}
