package block_validity_prover_balance_update_witness

import (
	"errors"
	"github.com/prodadidb/go-validation"
	intMaxAcc "intmax2-node/internal/accounts"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

// ErrValueLessOne error: must not be less than one.
var ErrValueLessOne = errors.New("must not be less than one")

// ErrMoreThenCurrentBlockNumber error: must not be more than the currentBlockNumber value.
var ErrMoreThenCurrentBlockNumber = errors.New("must not be more than the currentBlockNumber value")

// ErrLessThenTargetBlockNumber error: must not be less than the targetBlockNumber value.
var ErrLessThenTargetBlockNumber = errors.New("must not be less than the targetBlockNumber value")

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.User, validation.Required, input.IsUser()),
		validation.Field(&input.TargetBlockNumber,
			validation.Required,
			input.IsBlockNumber(),
			input.CheckIsMoreThenCurrentBlockNumber(),
			input.CheckIsTargetBlockNumberMoreThenCurrentBlockNumber(),
			input.CheckIsInvalidTargetBlockNumber(),
		),
		validation.Field(&input.CurrentBlockNumber,
			validation.Required, input.IsBlockNumber(),
			input.CheckIsLessThenTargetBlockNumber(),
			input.CheckIsCurrentBlockNumberLessThenTargetBlockNumber(),
			input.CheckIsInvalidCurrentBlockNumber(),
		),
	)
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) IsUser() validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		u, err := intMaxAcc.NewAddressFromHex(v)
		if err != nil {
			return ErrValueInvalid
		}

		_, err = u.Public()
		if err != nil {
			return ErrValueInvalid
		}

		return nil
	})
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) IsBlockNumber() validation.Rule {
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

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) CheckIsMoreThenCurrentBlockNumber() validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(int64)
		if !ok {
			return ErrValueInvalid
		}

		if v > input.CurrentBlockNumber {
			return ErrMoreThenCurrentBlockNumber
		}

		return nil
	})
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) CheckIsTargetBlockNumberMoreThenCurrentBlockNumber() validation.Rule {
	return validation.By(func(value interface{}) error {
		if input.IsTargetBlockNumberMoreThenCurrentBlockNumber {
			return ErrMoreThenCurrentBlockNumber
		}

		return nil
	})
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) CheckIsInvalidTargetBlockNumber() validation.Rule {
	return validation.By(func(_ interface{}) error {
		if input.IsInvalidTargetBlockNumber {
			return ErrValueInvalid
		}

		return nil
	})
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) CheckIsLessThenTargetBlockNumber() validation.Rule {
	return validation.By(func(value interface{}) error {
		v, ok := value.(int64)
		if !ok {
			return ErrValueInvalid
		}

		if v < input.TargetBlockNumber {
			return ErrLessThenTargetBlockNumber
		}

		return nil
	})
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) CheckIsCurrentBlockNumberLessThenTargetBlockNumber() validation.Rule {
	return validation.By(func(value interface{}) error {
		if input.IsCurrentBlockNumberLessThenTargetBlockNumber {
			return ErrLessThenTargetBlockNumber
		}

		return nil
	})
}

func (input *UCBlockValidityProverBalanceUpdateWitnessInput) CheckIsInvalidCurrentBlockNumber() validation.Rule {
	return validation.By(func(_ interface{}) error {
		if input.IsInvalidCurrentBlockNumber {
			return ErrValueInvalid
		}

		return nil
	})
}
