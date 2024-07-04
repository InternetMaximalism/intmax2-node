package models

import "time"

type WithdrawalGroupStatus int

const (
	PENDING WithdrawalGroupStatus = iota
	PROCESSING
	SUCCESS
	FAILED
)

type TransferMerkleProof struct {
	Index    int      `json:"index"`
	Siblings []string `json:"siblings"`
}

type Transaction struct {
	FeeTransferHash  string `json:"fee_transfer_hash"`
	TransferTreeRoot string `json:"transfer_tree_root"`
	TokenIndex       int    `json:"token_index"`
	Nonce            int    `json:"nonce"`
}

type TxMerkleProof struct {
	Index    int      `json:"index"`
	Siblings []string `json:"siblings"`
}

type Withdrawal struct {
	ID                  string              `json:"id"`
	Recipient           string              `json:"recipient"`
	TokenIndex          int                 `json:"token_index"`
	Amount              string              `json:"amount"`
	Salt                string              `json:"salt"`
	TransferHash        string              `json:"transfer_hash"`
	TransferMerkleProof TransferMerkleProof `json:"transfer_merkle_proof"`
	Transaction         Transaction         `json:"transaction"`
	TxMerkleProof       TxMerkleProof       `json:"tx_merkle_proof"`
	BlockNumber         int                 `json:"block_number"`
	EnoughBalanceProof  string              `json:"enough_balance_proof"`
	GroupID             *string             `json:"group_id"`
	CreatedAt           time.Time           `json:"created_at"`
}
