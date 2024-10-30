package get_backup_user_state

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCGetBackupUserStateInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.UserStateID, validation.Required),
	)
}
