package backup_deposit

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_deposit.go -package=mocks -source=get_backup_deposit.go

type UCGetBackupDepositContent struct {
	EncryptedDeposit string    `json:"encryptedDeposit"`
	BackupTime       time.Time `json:"backupTime"`
}

type UCGetBackupDepositMeta struct {
	StartBackupTime time.Time `json:"startBackupTime"`
	EndBackupTime   time.Time `json:"endBackupTime"`
}

type UCGetBackupDeposit struct {
	Deposits []UCGetBackupDepositContent `json:"deposits"`
	Meta     UCGetBackupDepositMeta      `json:"meta"`
}

type UCGetBackupDepositInput struct {
	Recipient       string               `json:"recipient"`
	DecodeRecipient *intMaxAcc.PublicKey `json:"-"`
	StartBackupTime uint32               `json:"blockNumber"`
	Limit           uint                 `json:"limit"`
}

// UseCaseGetBackupDeposit describes GetBackupDeposit contract.
type UseCaseGetBackupDeposit interface {
	Do(ctx context.Context, input *UCGetBackupDepositInput) (*UCGetBackupDeposit, error)
}
