package transaction

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_transaction.go -package=mocks -source=transaction.go

const (
	SuccessMsg   = "Transaction accepted and verified."
	NotUniqueMsg = "Transaction must be unique."
)

type RecipientTransferDataTransaction struct {
	AddressType string `json:"addressType"`
	Address     string `json:"address"`
}

type TransferDataTransaction struct {
	DecodeHash       *intMaxTypes.PoseidonHashOut      `json:"-"`
	Recipient        *RecipientTransferDataTransaction `json:"recipient"`
	DecodeRecipient  *intMaxTypes.GenericAddress       `json:"-"`
	TokenIndex       string                            `json:"tokenIndex"`
	DecodeTokenIndex *big.Int                          `json:"-"`
	Amount           string                            `json:"amount"`
	DecodeAmount     *big.Int                          `json:"-"`
	// signature
	Salt       string                       `json:"salt"`
	DecodeSalt *intMaxTypes.PoseidonHashOut `json:"-"`
}

type BackupTransferInput struct {
	Recipient                string `json:"recipient"`
	TransferHash             string `json:"transferHash"`
	EncodedEncryptedTransfer string `json:"encryptedTransfer"`
}

type UCTransactionInput struct {
	Sender             string                     `json:"sender"`
	DecodeSender       *intMaxAcc.PublicKey       `json:"-"`
	TransfersHash      string                     `json:"transfersHash"`
	Nonce              uint64                     `json:"nonce"` // initial value is 0
	PowNonce           string                     `json:"powNonce"`
	TransferData       []*TransferDataTransaction `json:"transferData"`
	DecodeTransferData []*intMaxTypes.Transfer    `json:"-"`
	Expiration         time.Time                  `json:"expiration"`
	Signature          string                     `json:"signature"`
}

// UseCaseTransaction describes Transaction contract.
type UseCaseTransaction interface {
	Do(ctx context.Context, input *UCTransactionInput) error
}
