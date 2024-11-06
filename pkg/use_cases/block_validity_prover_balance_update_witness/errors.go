package block_validity_prover_balance_update_witness

import "errors"

// ErrUCBlockValidityProverBalanceUpdateWitnessInputEmpty error: ucBlockValidityProverBalanceUpdateWitnessInput must not be empty.
var ErrUCBlockValidityProverBalanceUpdateWitnessInputEmpty = errors.New(
	"ucBlockValidityProverBalanceUpdateWitnessInput must not be empty",
)

// ErrNewAddressFromHexFail error: failed to create new address from hex.
var ErrNewAddressFromHexFail = errors.New("failed to create new address from hex")

// ErrPublicKeyFromIntMaxAccFail error: failed to get public key from INTMAX account.
var ErrPublicKeyFromIntMaxAccFail = errors.New("failed to get public key from INTMAX account")

// ErrFetchUpdateWitnessFail error: failed to fetch update witness.
var ErrFetchUpdateWitnessFail = errors.New("failed to fetch update witness")

// ErrCurrentBlockNumberLessThenTargetBlockNumber error: current block number must not be less then target block number.
var ErrCurrentBlockNumberLessThenTargetBlockNumber = errors.New(
	"current block number must not be less then target block number",
)

// ErrCurrentBlockNumberInvalid error: current block number must be valid.
var ErrCurrentBlockNumberInvalid = errors.New("current block number must be valid")

// ErrTargetBlockNumberInvalid error: current block number must be valid.
var ErrTargetBlockNumberInvalid = errors.New("target block number must be valid")
