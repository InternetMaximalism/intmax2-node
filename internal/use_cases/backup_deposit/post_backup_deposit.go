package backup_deposit

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_deposit.go -package=mocks -source=post_backup_deposit.go

type UCPostBackupDeposit struct {
	Message string `json:"message"`
}

type UCPostBackupDepositInput struct {
	Recipient        string               `json:"recipient"`
	DecodeRecipient  *intMaxAcc.PublicKey `json:"-"`
	BlockNumber      uint32               `json:"blockNumber"`
	EncryptedDeposit string               `json:"encryptedDeposit"`
}

// UseCasePostBackupDeposit describes PostBackupDeposit contract.
type UseCasePostBackupDeposit interface {
	Do(ctx context.Context, input *UCPostBackupDepositInput) (*UCPostBackupDeposit, error)
}
