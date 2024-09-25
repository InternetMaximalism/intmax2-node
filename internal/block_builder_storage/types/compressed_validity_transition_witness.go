package types

import (
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
)

type CompressedValidityTransitionWitness struct {
	SenderLeaves                         []SenderLeaf                       `json:"senderLeaves"`
	BlockMerkleProof                     intMaxTree.PoseidonMerkleProof     `json:"blockMerkleProof"`
	SignificantAccountRegistrationProofs *[]AccountRegistrationProofOrDummy `json:"significantAccountRegistrationProofs,omitempty"`
	SignificantAccountUpdateProofs       *[]intMaxTree.IndexedUpdateProof   `json:"significantAccountUpdateProofs,omitempty"`
	CommonAccountMerkleProof             []*intMaxGP.PoseidonHashOut        `json:"commonAccountMerkleProof"`
}
