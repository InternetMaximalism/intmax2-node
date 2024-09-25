package types

import intMaxTree "intmax2-node/internal/tree"

type AccountUpdateProofs struct {
	Proofs  []intMaxTree.IndexedUpdateProof `json:"proofs"`
	IsValid bool                            `json:"isValid"`
}

func (arp *AccountUpdateProofs) Set(other *AccountUpdateProofs) *AccountUpdateProofs {
	arp.IsValid = other.IsValid
	arp.Proofs = make([]intMaxTree.IndexedUpdateProof, len(other.Proofs))
	copy(arp.Proofs, other.Proofs)

	return arp
}
