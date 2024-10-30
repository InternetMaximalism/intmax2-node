package block_tree_proof_by_root_and_leaf_block_numbers

import (
	"errors"

	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

// ErrValueLessZero error: must not be less than zero.
var ErrValueLessZero = errors.New("must not be less than zero")

// ErrValueLessOne error: must not be less than one.
var ErrValueLessOne = errors.New("must not be less than one")

func (input *UCBlockTreeProofByRootAndLeafBlockNumbersInput) Valid() error {
	const (
		int0Key = 0
		int1Key = 1
	)

	return validation.ValidateStruct(input,
		validation.Field(&input.RootBlockNumber, validation.By(func(value interface{}) error {
			v, ok := value.(int64)
			if !ok {
				return ErrValueInvalid
			}

			if v < int1Key {
				return ErrValueLessOne
			}

			return nil
		})),
		validation.Field(&input.LeafBlockNumber, validation.By(func(value interface{}) error {
			v, ok := value.(int64)
			if !ok {
				return ErrValueInvalid
			}

			if v < int0Key {
				return ErrValueLessZero
			}

			return nil
		})),
	)
}
