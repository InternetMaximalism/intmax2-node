package balance_prover_service

import "errors"

// ErrBalanceProofNotGenerated error: balance proof is not generated.
var ErrBalanceProofNotGenerated = errors.New("balance proof is not generated")

// ErrStatusRequestTimeout error: get response with status code 408.
var ErrStatusRequestTimeout = errors.New("get response with status code 408")
