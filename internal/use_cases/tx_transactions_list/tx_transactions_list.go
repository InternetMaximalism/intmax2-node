package tx_transactions_list

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_tx_transactions_list.go -package=mocks -source=tx_transactions_list.go

type UCTxTransactionsListFilter struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

type UCTxTransactionsListPaginationCursor struct {
	BlockNumber  string `json:"blockNumber"`
	SortingValue string `json:"sortingValue"`
}

type UCTxTransactionsListPagination struct {
	Direction string                                `json:"direction"`
	Limit     string                                `json:"limit"`
	Cursor    *UCTxTransactionsListPaginationCursor `json:"cursor"`
}

type UCTxTransactionsListInput struct {
	Sorting    string                          `json:"sorting"`
	Pagination *UCTxTransactionsListPagination `json:"pagination"`
	Filter     *UCTxTransactionsListFilter     `json:"filter"`
}

type UseCaseTxTransactionsList interface {
	Do(ctx context.Context, input *UCTxTransactionsListInput, userEthPrivateKey string) (json.RawMessage, error)
}
