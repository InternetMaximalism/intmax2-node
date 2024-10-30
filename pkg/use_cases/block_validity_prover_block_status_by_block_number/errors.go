package block_validity_prover_block_status_by_block_number

import "errors"

// ErrBlockContentByBlockNumberFail error: failed to get block content by block number.
var ErrBlockContentByBlockNumberFail = errors.New("failed to get block content by block number")

// ErrRelationshipL2BatchIndexAndBlockContentsByBlockContentIDFail error: failed to get relationship l2-batch-index and block-contents by block-content-id.
var ErrRelationshipL2BatchIndexAndBlockContentsByBlockContentIDFail = errors.New(
	"failed to get relationship l2-batch-index and block-contents by block-content-id",
)

// ErrL2BatchIndexFail error: failed to get l2-batch-index info.
var ErrL2BatchIndexFail = errors.New("failed to get l2-batch-index info")
