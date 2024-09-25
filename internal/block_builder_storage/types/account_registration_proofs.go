package types

import intMaxTree "intmax2-node/internal/tree"

type AccountRegistrationProofs struct {
	Proofs  []intMaxTree.IndexedInsertionProof `json:"proofs"`
	IsValid bool                               `json:"isValid"`
}

func (arp *AccountRegistrationProofs) Set(other *AccountRegistrationProofs) *AccountRegistrationProofs {
	arp.IsValid = other.IsValid
	arp.Proofs = make([]intMaxTree.IndexedInsertionProof, len(other.Proofs))
	copy(arp.Proofs, other.Proofs)

	return arp
}
