package get_backup_transfers_list

import (
	"context"
	mFL "intmax2-node/internal/sql_filter/models"
	"math/big"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_transfers_list.go -package=mocks -source=get_backup_transfers_list.go

type UCGetBackupTransfersListCursorBase struct {
	BlockNumber         string   `json:"block_number"`
	ConvertBlockNumber  *big.Int `json:"-"`
	SortingValue        string   `json:"sorting_value"`
	ConvertSortingValue *big.Int `json:"-"`
}

type UCGetBackupTransfersListPaginationInput struct {
	Direction mFL.Direction                       `json:"direction"`
	PerPage   string                              `json:"per_page"`
	Offset    int                                 `json:"-"`
	Cursor    *UCGetBackupTransfersListCursorBase `json:"cursor"`
}

type UCGetBackupTransfersListInput struct {
	Recipient  string                                   `json:"recipient"`
	Pagination *UCGetBackupTransfersListPaginationInput `json:"pagination"`
	OrderBy    mFL.OrderBy                              `json:"order_by"`
	Sorting    mFL.Sorting                              `json:"sorting"`
	Filters    []*mFL.Filter                            `json:"filters"`
}

type UCGetBackupTransfersListCursorList struct {
	Prev *UCGetBackupTransfersListCursorBase `json:"prev"`
	Next *UCGetBackupTransfersListCursorBase `json:"next"`
}

type UCGetBackupTransfersListPaginationOfList struct {
	PerPage string                              `json:"per_page"`
	Cursor  *UCGetBackupTransfersListCursorList `json:"cursor"`
}

type ItemOfGetBackupTransfersList struct {
	ID                 string    `json:"id"`
	Recipient          string    `json:"recipient"`
	TransferDoubleHash string    `json:"transfer_double_hash"`
	EncryptedTransfer  string    `json:"encrypted_transfer"`
	BlockNumber        int64     `json:"block_number"`
	CreatedAt          time.Time `json:"created_at"`
}

type UCGetBackupTransfersList struct {
	Pagination UCGetBackupTransfersListPaginationOfList `json:"pagination"`
	List       []ItemOfGetBackupTransfersList           `json:"list"`
}

// UseCaseGetBackupTransfersList describes GetBackupTransfersList contract.
type UseCaseGetBackupTransfersList interface {
	Do(ctx context.Context, input *UCGetBackupTransfersListInput) (*UCGetBackupTransfersList, error)
}
