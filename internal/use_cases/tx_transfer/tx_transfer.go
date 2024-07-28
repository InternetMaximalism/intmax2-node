package tx_transfer

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_tx_transfer.go -package=mocks -source=tx_transfer.go

// type RecipientTransferDataTransaction struct {
// 	AddressType string `json:"addressType"`
// 	Address     string `json:"address"`
// }

// type TransferDataTransaction struct {
// 	DecodeHash       *intMaxTypes.PoseidonHashOut      `json:"-"`
// 	Recipient        *RecipientTransferDataTransaction `json:"recipient"`
// 	DecodeRecipient  *intMaxTypes.GenericAddress       `json:"-"`
// 	TokenIndex       string                            `json:"tokenIndex"`
// 	DecodeTokenIndex *big.Int                          `json:"-"`
// 	Amount           string                            `json:"amount"`
// 	DecodeAmount     *big.Int                          `json:"-"`
// 	// signature
// 	Salt       string                       `json:"salt"`
// 	DecodeSalt *intMaxTypes.PoseidonHashOut `json:"-"`
// }

type UseCaseTxTransfer interface {
	Do(ctx context.Context, args []string, amount, recipientAddressStr, userPrivateKey string) error
}
