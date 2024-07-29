package post_withdrawals_by_hashes

import (
	"context"
	mDBApp "intmax2-node/pkg/sql_db/db_app/models"
)

//go:generate mockgen -destination=../mocks/mock_post_withdrawals_by_hashes.go -package=mocks -source=post_withdrawals_by_hashes.go

const (
	SuccessMsg = "Withdrawal request accepted."
)

// Define the WithdrawalsByHashesRequest struct
type UCPostWithdrawalsByHashesInput struct {
	TransferHashes []string `json:"transfer_hashes"`
}

// UseCasePostWithdrawalsByHashes describes PostWithdrawalsByHashes
type UseCasePostWithdrawalsByHashes interface {
	Do(ctx context.Context, input *UCPostWithdrawalsByHashesInput) (*[]mDBApp.Withdrawal, error)
}
