package block_validity_prover_account

import (
	"errors"
	intMaxAcc "intmax2-node/internal/accounts"

	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCBlockValidityProverAccountInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.Address, validation.Required, input.IsAddress()),
	)
}

func (input *UCBlockValidityProverAccountInput) IsAddress() validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		addr, err := intMaxAcc.NewAddressFromHex(v)
		if err != nil {
			return ErrValueInvalid
		}

		_, err = addr.Public()
		if err != nil {
			return ErrValueInvalid
		}

		return nil
	})
}
