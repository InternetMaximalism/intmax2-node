package block_validity_prover_block_status_by_block_hash

import (
	"github.com/prodadidb/go-validation"
)

func (input *UCBlockValidityProverBlockStatusByBlockHashInput) Valid() error {
	return validation.ValidateStruct(input,
		validation.Field(&input.BlockHash, validation.Required),
	)
}
