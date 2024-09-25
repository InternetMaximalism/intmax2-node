package types

import intMaxTree "intmax2-node/internal/tree"

type MerkleTrees struct {
	AccountTree   *intMaxTree.AccountTree
	BlockHashTree *intMaxTree.BlockHashTree
	DepositLeaves []*intMaxTree.DepositLeaf
}
