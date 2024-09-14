package tree

import "errors"

// ErrInitSISiblings error: si siblings initializing error occurred.
var ErrInitSISiblings = errors.New("si siblings initializing error occurred")

// ErrBuildMerkleRootMoreThenOne error: in build Merkle root detected more than one root.
var ErrBuildMerkleRootMoreThenOne = errors.New("in build Merkle root detected more than one root")

// ErrInitSiblingsFail error: failed to initializing siblings with mt.count and initialLeaves length mismatch.
var ErrInitSiblingsFail = errors.New(
	"failed to initializing siblings with mt.count and initialLeaves length mismatch",
)

// ErrCalculateInitMerkelRootFail error: failed to calculating initial Merkle root.
var ErrCalculateInitMerkelRootFail = errors.New("failed to calculating initial Merkle root")

// ErrNewPoseidonMerkleTreeFail error: failed to create new Poseidon Merkle tree.
var ErrNewPoseidonMerkleTreeFail = errors.New("failed to create new Poseidon Merkle tree")

// ErrAddLeafFail error: failed to add leaf
var ErrAddLeafFail = errors.New("failed to add")

// ErrLeafInputIndexInvalid error: index is not equal to the length of leaves.
var ErrLeafInputIndexInvalid = errors.New("index is not equal to the length of leaves")

var ErrInvalidPrevRoot = errors.New("invalid prevRoot")

var ErrInvalidRootAfterUpdatedPrevLeaf = errors.New("invalid root after updated previous leaf")

// ErrAssetLeafInputIndexInvalid error: index is not equal to the length of asset leaves.
var ErrAssetLeafInputIndexInvalid = errors.New("index is not equal to the length of asset leaves")
