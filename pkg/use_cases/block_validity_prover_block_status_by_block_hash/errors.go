package block_validity_prover_block_status_by_block_hash

import "errors"

// ErrBlockContentByBlockHashFail error: failed to get block content by block hash.
var ErrBlockContentByBlockHashFail = errors.New("failed to get block content by block hash")

// ErrRelationshipL2BatchIndexAndBlockContentsByBlockContentIDFail error: failed to get relationship l2-batch-index and block-contents by block-content-id.
var ErrRelationshipL2BatchIndexAndBlockContentsByBlockContentIDFail = errors.New(
	"failed to get relationship l2-batch-index and block-contents by block-content-id",
)

// ErrL2BatchIndexFail error: failed to get l2-batch-index info.
var ErrL2BatchIndexFail = errors.New("failed to get l2-batch-index info")
