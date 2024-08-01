package withdrawal_request

import (
	"intmax2-node/configs"

	"github.com/prodadidb/go-validation"
)

func (input *UCWithdrawalInput) Valid(cfg *configs.Config) error {
	return validation.ValidateStruct(input)
}
