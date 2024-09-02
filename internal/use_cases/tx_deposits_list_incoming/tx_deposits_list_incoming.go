package tx_deposits_list_incoming

import (
	"context"
	"encoding/json"
)

//go:generate mockgen -destination=../mocks/mock_tx_deposits_list_incoming.go -package=mocks -source=tx_deposits_list_incoming.go

type UCTxDepositsListIncomingFilter struct {
	Name      string `json:"name"`
	Condition string `json:"condition"`
	Value     string `json:"value"`
}

type UCTxDepositsListIncomingPaginationCursor struct {
	BlockNumber  string `json:"blockNumber"`
	SortingValue string `json:"sortingValue"`
}

type UCTxDepositsListIncomingPagination struct {
	Direction string                                    `json:"direction"`
	Limit     string                                    `json:"limit"`
	Cursor    *UCTxDepositsListIncomingPaginationCursor `json:"cursor"`
}

type UCTxDepositsListIncomingInput struct {
	Sorting    string                              `json:"sorting"`
	Pagination *UCTxDepositsListIncomingPagination `json:"pagination"`
	Filter     *UCTxDepositsListIncomingFilter     `json:"filter"`
}

type UseCaseTxDepositsListIncoming interface {
	Do(ctx context.Context, input *UCTxDepositsListIncomingInput, userEthPrivateKey string) (json.RawMessage, error)
}
