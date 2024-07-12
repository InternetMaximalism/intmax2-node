package block_proposed

import "errors"

// ErrTxMerkleProofsByTxHashFail error: failed to get tx merkle proofs by hash.
var ErrTxMerkleProofsByTxHashFail = errors.New("failed to get tx merkle proofs by hash")

// ErrUnmarshalFail error: failed to unmarshal.
var ErrUnmarshalFail = errors.New("failed to unmarshal")

// ErrTrHashFail error: failed to get info by transfers hash.
var ErrTrHashFail = errors.New("failed to get info by transfers hash")
