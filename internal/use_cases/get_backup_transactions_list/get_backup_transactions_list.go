package get_backup_transactions_list

import (
	"context"
	mFL "intmax2-node/internal/sql_filter/models"
	"math/big"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transactions_list.go -package=mocks -source=get_backup_transactions_list.go

type UCGetBackupTransactionsListCursorBase struct {
	BlockNumber         string   `json:"block_number"`
	ConvertBlockNumber  *big.Int `json:"-"`
	SortingValue        string   `json:"sorting_value"`
	ConvertSortingValue *big.Int `json:"-"`
}

type UCGetBackupTransactionsListPaginationInput struct {
	Direction mFL.Direction                          `json:"direction"`
	PerPage   string                                 `json:"per_page"`
	Offset    int                                    `json:"-"`
	Cursor    *UCGetBackupTransactionsListCursorBase `json:"cursor"`
}

type UCGetBackupTransactionsListInput struct {
	Sender     string                                      `json:"sender"`
	Pagination *UCGetBackupTransactionsListPaginationInput `json:"pagination"`
	OrderBy    mFL.OrderBy                                 `json:"order_by"`
	Sorting    mFL.Sorting                                 `json:"sorting"`
	Filters    []*mFL.Filter                               `json:"filters"`
}

type UCGetBackupTransactionsListCursorList struct {
	Prev *UCGetBackupTransactionsListCursorBase `json:"prev"`
	Next *UCGetBackupTransactionsListCursorBase `json:"next"`
}

type UCGetBackupTransactionsListPaginationOfList struct {
	PerPage string                                 `json:"per_page"`
	Cursor  *UCGetBackupTransactionsListCursorList `json:"cursor"`
}

type ItemOfGetBackupTransactionsList struct {
	ID              string    `json:"id"`
	Sender          string    `json:"sender"`
	TxDoubleHash    string    `json:"tx_double_hash"`
	EncryptedTx     string    `json:"encrypted_tx"`
	EncodingVersion int64     `json:"encoding_version"`
	BlockNumber     int64     `json:"block_number"`
	Signature       string    `json:"signature"`
	CreatedAt       time.Time `json:"created_at"`
}

type UCGetBackupTransactionsList struct {
	Pagination UCGetBackupTransactionsListPaginationOfList `json:"pagination"`
	List       []ItemOfGetBackupTransactionsList           `json:"list"`
}

// UseCaseGetBackupTransactionsList describes GetBackupTransactionsList contract.
type UseCaseGetBackupTransactionsList interface {
	Do(ctx context.Context, input *UCGetBackupTransactionsListInput) (*UCGetBackupTransactionsList, error)
}
