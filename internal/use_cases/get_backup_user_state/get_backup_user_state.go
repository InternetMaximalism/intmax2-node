package get_backup_user_state

import (
	"context"
	"time"
)

//go:generate mockgen -destination=../mocks/mock_get_backup_user_state.go -package=mocks -source=get_backup_user_state.go

const (
	NotFoundMessage = "Backup user state not found."
	SuccessMsg      = "Backup user state found successful."
)

type UCGetBackupUserState struct {
	ID                 string
	UserAddress        string
	BalanceProof       string
	EncryptedUserState string
	AuthSignature      string
	BlockNumber        int64
	CreatedAt          time.Time
}

type UCGetBackupUserStateInput struct {
	UserStateID string `json:"userStateId"`
}

// UseCaseGetBackupUserState describes GetBackupUserState contract.
type UseCaseGetBackupUserState interface {
	Do(ctx context.Context, input *UCGetBackupUserStateInput) (*UCGetBackupUserState, error)
}
