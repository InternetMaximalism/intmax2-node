package post_backup_transfer

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_transfer.go -package=mocks -source=post_backup_transfer.go

const (
	SuccessMsg = "Backup transfer accepted."
)

type UCPostBackupTransferInput struct {
	TransferHash               string `json:"transferHash"`
	EncryptedTransfer          string `json:"encryptedTransfer"`
	SenderLastBalanceProofBody string `json:"senderLastBalanceProofBody"`
	SenderTransitionProofBody  string `json:"senderBalanceTransitionProofBody"`
	Recipient                  string `json:"recipient"`
	BlockNumber                uint32 `json:"blockNumber"`
}

// UseCasePostBackupTransfer describes PostBackupTransfer contract.
type UseCasePostBackupTransfer interface {
	Do(ctx context.Context, input *UCPostBackupTransferInput) error
}
