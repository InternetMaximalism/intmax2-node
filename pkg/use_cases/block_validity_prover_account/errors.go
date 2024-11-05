package block_validity_prover_account

import "errors"

// ErrUCBlockValidityProverAccountInputEmpty error: ucBlockValidityProverAccountInput must not be empty.
var ErrUCBlockValidityProverAccountInputEmpty = errors.New(
	"ucBlockValidityProverAccountInput must not be empty",
)

// ErrNewAddressFromHexFail error: failed to create new address from hex.
var ErrNewAddressFromHexFail = errors.New("failed to create new address from hex")

// ErrSenderByAddressFail error: failed to get sender by address.
var ErrSenderByAddressFail = errors.New("failed to get sender by address")

// ErrAccountBySenderIDFail error: failed to get account by sender ID.
var ErrAccountBySenderIDFail = errors.New("failed to get account by sender ID")
