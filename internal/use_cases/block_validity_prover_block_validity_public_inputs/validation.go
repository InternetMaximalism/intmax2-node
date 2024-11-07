package block_validity_prover_block_validity_public_inputs

import (
	"errors"

	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

// ErrValueLessOne error: must not be less than one.
var ErrValueLessOne = errors.New("must not be less than one")

func (input *UCBlockValidityProverBlockValidityPublicInputsInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.BlockNumber,
			validation.Required,
			input.IsBlockNumber(),
		),
	)
}

func (input *UCBlockValidityProverBlockValidityPublicInputsInput) IsBlockNumber() validation.Rule {
	const int1Key = 1

	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(int64)
		if !ok {
			return ErrValueInvalid
		}

		if v < int1Key {
			return ErrValueLessOne
		}

		return nil
	})
}
