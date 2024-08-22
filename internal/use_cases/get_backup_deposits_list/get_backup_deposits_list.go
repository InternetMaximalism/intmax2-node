package get_backup_deposits_list

import (
	"context"
	mFL "intmax2-node/internal/sql_filter/models"
	"math/big"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_deposits_list.go -package=mocks -source=get_backup_deposits_list.go

type UCGetBackupDepositsListCursorBase struct {
	BlockNumber         string   `json:"block_number"`
	ConvertBlockNumber  *big.Int `json:"-"`
	SortingValue        string   `json:"sorting_value"`
	ConvertSortingValue *big.Int `json:"-"`
}

type UCGetBackupDepositsListPaginationInput struct {
	Direction mFL.Direction                      `json:"direction"`
	PerPage   string                             `json:"per_page"`
	Offset    int                                `json:"-"`
	Cursor    *UCGetBackupDepositsListCursorBase `json:"cursor"`
}

type UCGetBackupDepositsListInput struct {
	Recipient  string                                  `json:"recipient"`
	Pagination *UCGetBackupDepositsListPaginationInput `json:"pagination"`
	OrderBy    mFL.OrderBy                             `json:"order_by"`
	Sorting    mFL.Sorting                             `json:"sorting"`
	Filters    []*mFL.Filter                           `json:"filters"`
}

type UCGetBackupDepositsListCursorList struct {
	Prev *UCGetBackupDepositsListCursorBase `json:"prev"`
	Next *UCGetBackupDepositsListCursorBase `json:"next"`
}

type UCGetBackupDepositsListPaginationOfList struct {
	PerPage string                             `json:"per_page"`
	Cursor  *UCGetBackupDepositsListCursorList `json:"cursor"`
}

type ItemOfGetBackupDepositsList struct {
	ID                string    `json:"id"`
	Recipient         string    `json:"recipient"`
	DepositDoubleHash string    `json:"deposit_double_hash"`
	EncryptedDeposit  string    `json:"encrypted_deposit"`
	BlockNumber       int64     `json:"block_number"`
	CreatedAt         time.Time `json:"created_at"`
}

type UCGetBackupDepositsList struct {
	Pagination UCGetBackupDepositsListPaginationOfList `json:"pagination"`
	List       []ItemOfGetBackupDepositsList           `json:"list"`
}

// UseCaseGetBackupDepositsList describes GetBackupDepositsList contract.
type UseCaseGetBackupDepositsList interface {
	Do(ctx context.Context, input *UCGetBackupDepositsListInput) (*UCGetBackupDepositsList, error)
}
