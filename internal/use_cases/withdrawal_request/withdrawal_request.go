package withdrawal_request

import (
	intMaxAcc "intmax2-node/internal/accounts"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"
)

//go:generate mockgen -destination=../mocks/mock_withdrawal_request.go -package=mocks -source=withdrawal_request.go

const (
	SuccessMsg   = "Transaction accepted and verified."
	NotUniqueMsg = "Transaction must be unique."
)

type UCWithdrawalInput struct {
	TransferData        *TransferDataTransaction `json:"transferData"`
	DecodeTransferData  *intMaxTypes.Transfer    `json:"-"`
	TransferMerkleProof TransferMerkleProof      `json:"transferMerkleProof"`
	Transaction         Transaction              `json:"transaction"`
	TxMerkleProof       TxMerkleProof            `json:"txMerkleProof"`
	TransferHash        string                   `json:"transferHash"`
	BlockNumber         uint32                   `json:"blockNumber"`
	BlockHash           string                   `json:"blockHash"`
	EnoughBalanceProof  EnoughBalanceProof       `json:"enoughBalanceProof"`
}

type TransferDataTransaction struct {
	DecodeHash       *intMaxTypes.PoseidonHashOut `json:"-"`
	Recipient        string                       `json:"recipient"`
	DecodeRecipient  *intMaxAcc.Address           `json:"-"`
	TokenIndex       string                       `json:"tokenIndex"`
	DecodeTokenIndex *big.Int                     `json:"-"`
	Amount           string                       `json:"amount"`
	DecodeAmount     *big.Int                     `json:"-"`
	// signature
	Salt       string                       `json:"salt"`
	DecodeSalt *intMaxTypes.PoseidonHashOut `json:"-"`
}

type TransferMerkleProof struct {
	Siblings []string `json:"siblings"`
	Index    int32    `json:"index"`
}

type Transaction struct {
	TransferTreeRoot string `json:"transferTreeRoot"`
	Nonce            int32  `json:"nonce"`
}

type TxMerkleProof struct {
	Siblings []string `json:"siblings"`
	Index    int32    `json:"index"`
}

type EnoughBalanceProof struct {
	Proof        string `json:"proof"`
	PublicInputs string `json:"publicInputs"`
}
