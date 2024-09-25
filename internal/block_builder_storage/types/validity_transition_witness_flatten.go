package types

import intMaxTree "intmax2-node/internal/tree"

type ValidityTransitionWitnessFlatten struct {
	SenderLeaves              []SenderLeaf                       `json:"senderLeaves"`
	BlockMerkleProof          intMaxTree.PoseidonMerkleProof     `json:"blockMerkleProof"`
	AccountRegistrationProofs []intMaxTree.IndexedInsertionProof `json:"accountRegistrationProofs,omitempty"`
	AccountUpdateProofs       []intMaxTree.IndexedUpdateProof    `json:"accountUpdateProofs,omitempty"`
}
