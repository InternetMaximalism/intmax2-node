package tx_withdrawal_transfers_list

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_tx_withdrawal_transfers_list.go -package=mocks -source=tx_withdrawal_transfers_list.go

type UCTxWithdrawalTransfersListFilter struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

type UCTxWithdrawalTransfersListPaginationCursor struct {
	BlockNumber  string `json:"blockNumber"`
	SortingValue string `json:"sortingValue"`
}

type UCTxWithdrawalTransfersListPagination struct {
	Direction string                                       `json:"direction"`
	Limit     string                                       `json:"limit"`
	Cursor    *UCTxWithdrawalTransfersListPaginationCursor `json:"cursor"`
}

type UCTxWithdrawalTransfersListInput struct {
	Sorting    string                                 `json:"sorting"`
	Pagination *UCTxWithdrawalTransfersListPagination `json:"pagination"`
	Filter     *UCTxWithdrawalTransfersListFilter     `json:"filter"`
}

type UseCaseTxWithdrawalTransfersList interface {
	Do(ctx context.Context, input *UCTxWithdrawalTransfersListInput, userEthPrivateKey string) (json.RawMessage, error)
}
