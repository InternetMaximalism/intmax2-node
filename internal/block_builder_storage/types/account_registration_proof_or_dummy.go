package types

import intMaxTree "intmax2-node/internal/tree"

type AccountRegistrationProofOrDummy struct {
	LowLeafProof *intMaxTree.PoseidonMerkleProof `json:"lowLeafProof,omitempty"`
	LeafProof    *intMaxTree.PoseidonMerkleProof `json:"leafProof,omitempty"`
	Index        uint64                          `json:"index"`
	LowLeafIndex uint64                          `json:"lowLeafIndex"`
	PrevLowLeaf  intMaxTree.IndexedMerkleLeaf    `json:"prevLowLeaf"`
}
