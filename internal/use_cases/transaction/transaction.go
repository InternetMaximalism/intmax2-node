package transaction

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/hash/goldenposeidon"
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

type BackupTransactionData struct {
	TxHash             string `json:"txHash"`
	EncodedEncryptedTx string `json:"encryptedTx"`
	Signature          string `json:"signature"`
}

func NewBackupTransactionData(
	userPublicKey *intMaxAcc.PublicKey,
	txDetails intMaxTypes.TxDetails,
	txHash *goldenposeidon.PoseidonHashOut,
	signature string,
) (*BackupTransactionData, error) {
	encodedTx := txDetails.Marshal()
	encryptedTx, err := intMaxAcc.EncryptECIES(
		rand.Reader,
		userPublicKey,
		encodedTx,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt deposit: %w", err)
	}

	encodedEncryptedTx := base64.StdEncoding.EncodeToString(encryptedTx)

	return &BackupTransactionData{
		TxHash:             txHash.String(),
		EncodedEncryptedTx: encodedEncryptedTx,
		Signature:          signature,
	}, nil
}

type BackupTransferInput struct {
	Recipient                  string `json:"recipient"`
	TransferHash               string `json:"transferHash"`
	EncodedEncryptedTransfer   string `json:"encryptedTransfer"`
	SenderLastBalanceProofBody string `json:"senderLastBalanceProofBody"`
	SenderTransitionProofBody  string `json:"senderBalanceTransitionProofBody"`
}

type UCTransactionInput struct {
	Sender             string                     `json:"sender"`
	DecodeSender       *intMaxAcc.PublicKey       `json:"-"`
	TransfersHash      string                     `json:"transfersHash"`
	Nonce              uint64                     `json:"nonce"`
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
