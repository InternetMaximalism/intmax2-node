package post_backup_transaction

import (
	"context"
	"intmax2-node/internal/use_cases/block_signature"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_transaction.go -package=mocks -source=post_backup_transaction.go

const (
	SuccessMsg = "Backup transaction accepted."
)

type UCPostBackupTransactionInput struct {
	TxHash                       string                                       `json:"txHash"`
	EncryptedTx                  string                                       `json:"encryptedTx"`
	SenderEnoughBalanceProofBody *block_signature.EnoughBalanceProofBodyInput `json:"senderEnoughBalanceProofBody"`
	Sender                       string                                       `json:"sender"`
	BlockNumber                  uint32                                       `json:"blockNumber"`
	Signature                    string                                       `json:"signature"`
	// SenderLastBalanceProofBody       []byte `json:"senderLastBalanceProofBody"`
	// SenderBalanceTransitionProofBody []byte `json:"senderBalanceTransitionProofBody"`
}

// UseCasePostBackupTransaction describes PostBackupTransaction contract.
type UseCasePostBackupTransaction interface {
	Do(ctx context.Context, input *UCPostBackupTransactionInput) (senderEnoughBalanceProofBodyHash string, err error)
}
