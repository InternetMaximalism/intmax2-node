package block_validity_prover_deposits

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/prodadidb/go-validation"
)

// ErrValueInvalid error: value must be valid.
var ErrValueInvalid = errors.New("must be a valid value")

func (input *UCBlockValidityProverDepositsInput) Valid() error {
	const (
		int1Key   = 1
		int100Key = 100
	)

	return validation.ValidateStruct(input,
		validation.Field(&input.DepositHashes,
			validation.Required,
			validation.Length(int1Key, int100Key),
			validation.Each(input.IsDepositHash()),
		),
	)
}

func (input *UCBlockValidityProverDepositsInput) IsDepositHash() validation.Rule {
	return validation.By(func(value interface{}) (err error) {
		v, ok := value.(string)
		if !ok {
			return ErrValueInvalid
		}

		var t common.Hash
		err = t.Scan(common.FromHex(v))
		if err != nil {
			return ErrValueInvalid
		}

		input.ConvertDepositHashes = append(input.ConvertDepositHashes, t)

		return nil
	})
}
