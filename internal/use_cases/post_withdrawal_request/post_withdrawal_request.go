package post_withdrawal_request

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_post_withdrawal_request.go -package=mocks -source=post_withdrawal_request.go

const (
	SuccessMsg = "Withdraw request accepted."
)

type UCPostWithdrawalRequestTransferDataInput struct {
	Recipient  string `json:"recipient"`
	TokenIndex int64  `json:"tokenIndex"`
	Amount     string `json:"amount"`
	Salt       string `json:"salt"`
}

type UCPostWithdrawalRequestTransferMerkleProofInput struct {
	Siblings []string `json:"siblings"`
	Index    int64    `json:"index"`
}

type UCPostWithdrawalRequestTransactionInput struct {
	TransferTreeRoot string `json:"transfer_tree_root"`
	Nonce            int64  `json:"nonce"`
}

type UCPostWithdrawalRequestTxMerkleProofInput struct {
	Siblings []string `json:"siblings"`
	Index    int64    `json:"index"`
}

type UCPostWithdrawalRequestEnoughBalanceProofInput struct {
	Proof        string `json:"proof"`
	PublicInputs string `json:"public_inputs"`
}

type UCPostWithdrawalRequestInput struct {
	TransferData        *UCPostWithdrawalRequestTransferDataInput        `json:"transferData"`
	TransferMerkleProof *UCPostWithdrawalRequestTransferMerkleProofInput `json:"transferMerkleProof"`
	Transaction         *UCPostWithdrawalRequestTransactionInput         `json:"transaction"`
	TxMerkleProof       *UCPostWithdrawalRequestTxMerkleProofInput       `json:"txMerkleProof"`
	TransferHash        string                                           `json:"transferHash"`
	BlockNumber         int64                                            `json:"blockNumber"`
	BlockHash           string                                           `json:"blockHash"`
	EnoughBalanceProof  *UCPostWithdrawalRequestEnoughBalanceProofInput  `json:"enoughBalanceProof"`
}

/**


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
*/

// UseCasePostWithdrawalRequest describes PostWithdrawalRequest
type UseCasePostWithdrawalRequest interface {
	Do(ctx context.Context, input *UCPostWithdrawalRequestInput) error
}
