package models

import "time"

type WithdrawalStatus int

const (
	WS_PENDING WithdrawalStatus = iota
	WS_SUCCESS
	WS_FAILED
)

func (s WithdrawalStatus) String() string {
	return [...]string{"pending", "success", "failed"}[s]
}

type TransferData struct {
	Recipient  string `json:"recipient"`
	TokenIndex int64  `json:"token_index"`
	Amount     string `json:"amount"`
	Salt       string `json:"salt"`
}

type TransferMerkleProof struct {
	Siblings []string `json:"siblings"`
	Index    int64    `json:"index"`
}

type Transaction struct {
	TransferTreeRoot string `json:"transfer_tree_root"`
	Nonce            int64  `json:"nonce"`
}

type TxMerkleProof struct {
	Siblings []string `json:"siblings"`
	Index    int64    `json:"index"`
}

type EnoughBalanceProof struct {
	Proof        string `json:"proof"`
	PublicInputs string `json:"public_inputs"`
}

type Withdrawal struct {
	ID                  string              `json:"id"`
	Status              int64               `json:"status"`
	TransferData        TransferData        `json:"transfer_data"`
	TransferMerkleProof TransferMerkleProof `json:"transfer_merkle_proof"`
	Transaction         Transaction         `json:"transaction"`
	TxMerkleProof       TxMerkleProof       `json:"tx_merkle_proof"`
	TransferHash        string              `json:"transfer_hash"`
	BlockNumber         int64               `json:"block_number"`
	BlockHash           string              `json:"block_hash"`
	EnoughBalanceProof  EnoughBalanceProof  `json:"enough_balance_proof"`
	CreatedAt           time.Time           `json:"created_at"`
}
