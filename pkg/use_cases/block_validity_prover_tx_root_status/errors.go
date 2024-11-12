package block_validity_prover_tx_root_status

import "errors"

// ErrUCBlockValidityProverTxRootStatusInputEmpty error: ucBlockValidityProverTxRootStatusInput must not be empty.
var ErrUCBlockValidityProverTxRootStatusInputEmpty = errors.New(
	"ucBlockValidityProverTxRootStatusInput must not be empty",
)
