package block_validity_prover_deposits

import "errors"

// ErrUCBlockValidityProverDepositsInputEmpty error: ucBlockValidityProverDepositsInput must not be empty.
var ErrUCBlockValidityProverDepositsInputEmpty = errors.New(
	"ucBlockValidityProverDepositsInput must not be empty",
)
