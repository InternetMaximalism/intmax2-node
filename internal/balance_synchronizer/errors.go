package balance_synchronizer

import "errors"

// ErrValidSentTxFail error: failed to valid sent tx.
var ErrValidSentTxFail = errors.New("failed to validate sent tx")

// ErrValidReceivedDepositFail error: failed to validate received deposit.
var ErrValidReceivedDepositFail = errors.New("failed to validate received deposit")

// ErrValidReceivedTransferFail error: failed to validate received transfer.
var ErrValidReceivedTransferFail = errors.New("failed to validate received transfer")

// ErrApplyReceivedDepositTransitionFail error: failed to apply received deposit transition.
var ErrApplyReceivedDepositTransitionFail = errors.New("failed to apply received deposit transition")

// ErrApplyReceivedTransferTransitionFail error: failed to apply received transfer transition.
var ErrApplyReceivedTransferTransitionFail = errors.New("failed to apply received transfer transition")

// ErrNewBalanceTransitionDataFail error: failed to start Balance Prover Service.
var ErrNewBalanceTransitionDataFail = errors.New("failed to start Balance Prover Service")

// ErrSortValidUserDataFail error: failed to sort valid user data.
var ErrSortValidUserDataFail = errors.New("failed to sort valid user data")

// ErrLatestSynchronizedBlockNumberFail error: failed to get latest synchronized block number.
var ErrLatestSynchronizedBlockNumberFail = errors.New("failed to get latest synchronized block number")

// ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber error: latest synchronized block number must be more last updated block number.
var ErrLatestSynchronizedBlockNumberLassOrEqualLastUpdatedBlockNumber = errors.New(
	"latest synchronized block number must be more last updated block number",
)

// ErrReceiveDepositAndUpdate error: failed to receive deposit and update
var ErrReceiveDepositAndUpdate = errors.New("failed to receive deposit and update")

// ErrNewCompressedPlonky2ProofFromBase64StringFail error: failed to create new compressed plonky2 proof from base64 string
var ErrNewCompressedPlonky2ProofFromBase64StringFail = errors.New("failed to create new compressed plonky2 proof from base64 string")

var ErrBalancePublicInputsFromPublicInputs = errors.New("failed to create new balance public inputs from public inputs")
var ErrProveReceiveDeposit = errors.New("failed to prove receive deposit")

// ErrNoValidUserData error: no valid user data
var ErrNoValidUserData = errors.New("no valid user data")
