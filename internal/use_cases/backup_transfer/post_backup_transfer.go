package backup_transfer

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_transfer.go -package=mocks -source=post_backup_transfer.go

type UCPostBackupTransfer struct {
	Message string `json:"message"`
}

type UCPostBackupTransferInput struct {
	Recipient         string               `json:"recipient"`
	DecodeRecipient   *intMaxAcc.PublicKey `json:"-"`
	BlockNumber       uint32               `json:"blockNumber"`
	EncryptedTransfer string               `json:"encryptedTransfer"`
}

// UseCasePostBackupTransfer describes PostBackupTransfer contract.
type UseCasePostBackupTransfer interface {
	Do(ctx context.Context, input *UCPostBackupTransferInput) (*UCPostBackupTransfer, error)
}
