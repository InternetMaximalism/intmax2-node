package types

import intMaxTree "intmax2-node/internal/tree"

type UpdateWitness struct {
	ValidityProof          string                             `json:"validityProof"`
	BlockMerkleProof       intMaxTree.BlockHashMerkleProof    `json:"blockMerkleProof"`
	AccountMembershipProof *intMaxTree.IndexedMembershipProof `json:"accountMembershipProof"`
}
