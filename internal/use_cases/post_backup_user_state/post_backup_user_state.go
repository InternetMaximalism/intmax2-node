package post_backup_user_state

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_post_backup_user_state.go -package=mocks -source=post_backup_user_state.go

const (
	SuccessMsg       = "Backup user state accepted."
	AlreadyExistsMsg = "Backup user state already exists."
)

type UCPostBackupUserState struct {
	ID                 string
	UserAddress        string
	BalanceProof       string
	EncryptedUserState string
	AuthSignature      string
	BlockNumber        int64
	CreatedAt          time.Time
}

type UCPostBackupUserStateInput struct {
	UserAddress        string `json:"userAddress"`
	BalanceProof       string `json:"balanceProof"`
	EncryptedUserState string `json:"encryptedUserState"`
	AuthSignature      string `json:"authSignature"`
	BlockNumber        int64  `json:"blockNumber"`
}

// UseCasePostBackupUserState describes PostBackupUserState contract.
type UseCasePostBackupUserState interface {
	Do(ctx context.Context, input *UCPostBackupUserStateInput) (*UCPostBackupUserState, error)
}
