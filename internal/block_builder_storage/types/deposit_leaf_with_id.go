package types

import intMaxTree "intmax2-node/internal/tree"

type DepositLeafWithId struct {
	DepositLeaf *intMaxTree.DepositLeaf
	DepositId   uint32
}
