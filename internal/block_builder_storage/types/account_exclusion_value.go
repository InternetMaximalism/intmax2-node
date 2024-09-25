package types

import (
	"errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type AccountExclusionValue struct {
	AccountTreeRoot         *intMaxGP.PoseidonHashOut
	AccountMembershipProofs []intMaxTree.IndexedMembershipProof
	PublicKeys              []intMaxTypes.Uint256
	PublicKeyCommitment     *intMaxGP.PoseidonHashOut
	IsValid                 bool
}

func NewAccountExclusionValue(
	accountTreeRoot *intMaxGP.PoseidonHashOut,
	accountMembershipProofs []intMaxTree.IndexedMembershipProof,
	publicKeys []intMaxTypes.Uint256,
) (*AccountExclusionValue, error) {
	result := true
	for i, proof := range accountMembershipProofs {
		err := proof.Verify(publicKeys[i].BigInt(), accountTreeRoot)
		if err != nil {
			var ErrAccountMembershipProofInvalid = errors.New("account membership proof is invalid")
			return nil, errors.Join(ErrAccountMembershipProofInvalid, err)
		}

		isDummy := publicKeys[i].IsDummyPublicKey()
		isExcluded := !proof.IsIncluded || isDummy
		result = result && isExcluded
	}

	publicKeyCommitment := GetPublicKeyCommitment(publicKeys)

	return &AccountExclusionValue{
		AccountTreeRoot:         accountTreeRoot,
		AccountMembershipProofs: accountMembershipProofs,
		PublicKeys:              publicKeys,
		PublicKeyCommitment:     publicKeyCommitment,
		IsValid:                 result,
	}, nil
}
