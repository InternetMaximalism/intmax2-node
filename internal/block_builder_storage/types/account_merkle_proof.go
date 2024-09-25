package types

import (
	"errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type AccountMerkleProof struct {
	MerkleProof intMaxTree.IndexedMerkleProof `json:"merkleProof"`
	Leaf        intMaxTree.IndexedMerkleLeaf  `json:"leaf"`
}

func (proof *AccountMerkleProof) Verify(publicKey intMaxTypes.Uint256, accountID uint64, accountTreeRoot *intMaxGP.PoseidonHashOut) error {
	if publicKey.IsDummyPublicKey() {
		return errors.New("public key is zero")
	}

	err := proof.MerkleProof.Verify(&proof.Leaf, int(accountID), accountTreeRoot)
	if err != nil {
		var ErrMerkleProofInvalid = errors.New("given Merkle proof is invalid")
		return errors.Join(ErrMerkleProofInvalid, err)
	}

	if publicKey.BigInt().Cmp(proof.Leaf.Key) != 0 {
		return errors.New("public key does not match leaf key")
	}

	return nil
}
