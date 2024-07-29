package backup_deposit

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_deposit.go -package=mocks -source=post_backup_deposit.go

const (
	SuccessMsg = "Backup deposit accepted."
)

type UCPostBackupDepositInput struct {
	Recipient        string `json:"recipient"`
	EncryptedDeposit string `json:"encryptedDeposit"`
	BlockNumber      uint32 `json:"blockNumber"`
}

// UseCasePostBackupDeposit describes PostBackupDeposit contract.
type UseCasePostBackupDeposit interface {
	Do(ctx context.Context, input *UCPostBackupDepositInput) error
}
