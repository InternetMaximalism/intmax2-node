package backup_transfer

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transfer.go -package=mocks -source=get_backup_transfer.go

type UCGetBackupTransferContent struct {
	EncryptedTransfer uint32    `json:"encryptedTransfer"`
	BackupTime        time.Time `json:"backupTime"`
}

type UCGetBackupTransferMeta struct {
	StartBackupTime time.Time `json:"startBackupTime"`
	EndBackupTime   time.Time `json:"endBackupTime"`
}

type UCGetBackupTransfer struct {
	Transfers []UCGetBackupTransferContent `json:"transfers"`
	Meta      UCGetBackupTransferMeta      `json:"meta"`
}

type UCGetBackupTransferInput struct {
	Recipient       string               `json:"recipient"`
	DecodeRecipient *intMaxAcc.PublicKey `json:"-"`
	StartBackupTime uint32               `json:"startBackupTime"`
	Limit           uint                 `json:"limit"`
}

// UseCaseGetBackupTransfer describes GetBackupTransfer contract.
type UseCaseGetBackupTransfer interface {
	Do(ctx context.Context, input *UCGetBackupTransferInput) (*UCGetBackupTransfer, error)
}
