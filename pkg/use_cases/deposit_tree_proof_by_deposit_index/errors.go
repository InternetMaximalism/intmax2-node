package deposit_tree_proof_by_deposit_index

import "errors"

// ErrUCDepositTreeProofByDepositIndexInputEmpty error: ucDepositTreeProofByDepositIndexInput must not be empty.
var ErrUCDepositTreeProofByDepositIndexInputEmpty = errors.New(
	"ucDepositTreeProofByDepositIndexInput must not be empty",
)

// ErrDepositTreeProofFail error: failed to get deposit tree proof.
var ErrDepositTreeProofFail = errors.New("failed to get deposit tree proof")
