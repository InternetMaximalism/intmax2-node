package types

import intMaxTree "intmax2-node/internal/tree"

type DepositInfo struct {
	DepositId      uint32
	DepositIndex   *uint32
	BlockNumber    *uint32
	IsSynchronized bool
	DepositLeaf    *intMaxTree.DepositLeaf
}
