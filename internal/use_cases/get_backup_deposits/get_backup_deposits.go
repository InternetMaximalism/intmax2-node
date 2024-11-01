package get_backup_deposits

//go:generate mockgen -destination=../mocks/mock_get_backup_deposits.go -package=mocks -source=get_backup_deposits.go

type UCGetBackupDepositsInput struct {
	Sender           string `json:"sender"`
	StartBlockNumber uint64 `json:"startBlockNumber"`
	Limit            uint64 `json:"limit"`
}

// // UseCaseGetBackupDeposits describes GetBackupDeposits contract.
// type UseCaseGetBackupDeposits interface {
// 	Do(ctx context.Context, input *UCGetBackupDepositsInput) (*node.GetBackupDepositsResponse_Data, error)
// }
