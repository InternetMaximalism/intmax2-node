package verify_deposit_confirmation

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_get_verify_deposit_confirmation.go -package=mocks -source=get_verify_deposit_confirmation.go

type UCGetVerifyDepositConfirmationInput struct {
	DepositId string `json:"deposit_id"`
}

// UseCaseGetVerifyDepositConfirmation describes GetVerifyDepositConfirmation contract.
type UseCaseGetVerifyDepositConfirmation interface {
	Do(ctx context.Context, input *UCGetVerifyDepositConfirmationInput) (bool, error)
}
