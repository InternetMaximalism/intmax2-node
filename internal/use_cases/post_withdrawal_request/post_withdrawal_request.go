package post_withdrawal_request

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=../mocks/mock_post_withdrawal_request.go -package=mocks -source=post_withdrawal_request.go

const (
	SuccessMsg = "Withdrawal request accepted."
)

// Define the WithdrawalRequestRequest struct
type UCPostWithdrawalRequestInput struct {
	TransferData        mDBApp.TransferData        `json:"transfer_data"`
	TransferMerkleProof mDBApp.TransferMerkleProof `json:"transfer_merkle_proof"`
	Transaction         mDBApp.Transaction         `json:"transaction"`
	TxMerkleProof       mDBApp.TxMerkleProof       `json:"tx_merkle_proof"`
	TransferHash        string                     `json:"transfer_hash"`
	BlockNumber         uint64                     `json:"block_number"`
	BlockHash           string                     `json:"block_hash"`
	EnoughBalanceProof  mDBApp.EnoughBalanceProof  `json:"enough_balance_proof"`
}

// UseCasePostWithdrawalRequest describes PostWithdrawalRequest
type UseCasePostWithdrawalRequest interface {
	Do(ctx context.Context, input *UCPostWithdrawalRequestInput) error
}
