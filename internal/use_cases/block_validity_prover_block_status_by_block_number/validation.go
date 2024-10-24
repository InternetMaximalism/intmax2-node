package block_validity_prover_block_status_by_block_number

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCBlockValidityProverBlockStatusByBlockNumberInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.BlockNumber, validation.Required),
	)
}
