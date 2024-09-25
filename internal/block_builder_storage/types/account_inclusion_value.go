package types

import (
	"errors"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type AccountInclusionValue struct {
	AccountIDPacked     AccountIdPacked
	AccountIDHash       intMaxTypes.Bytes32
	AccountTreeRoot     *intMaxGP.PoseidonHashOut
	AccountMerkleProofs []AccountMerkleProof
	PublicKeys          []intMaxTypes.Uint256
	PublicKeyCommitment *intMaxGP.PoseidonHashOut
	IsValid             bool
}

func NewAccountInclusionValue(
	accountTreeRoot *intMaxTree.PoseidonHashOut,
	accountIDPacked *AccountIdPacked,
	accountMerkleProofs []AccountMerkleProof,
	publicKeys []intMaxTypes.Uint256,
) (*AccountInclusionValue, error) {
	if len(accountMerkleProofs) != intMaxTypes.NumOfSenders {
		return nil, errors.New("account merkle proofs length should be equal to number of senders")
	}

	if len(publicKeys) != intMaxTypes.NumOfSenders {
		return nil, errors.New("public keys length should be equal to number of senders")
	}

	result := true
	accountIDHash := accountIDPacked.Hash()
	accountIDs := accountIDPacked.Unpack()
	for i := range accountIDs {
		accountID := accountIDs[i]
		proof := accountMerkleProofs[i]
		publicKey := publicKeys[i]
		err := proof.Verify(publicKey, accountID, accountTreeRoot)
		result = result && err == nil
	}

	publicKeyCommitment := GetPublicKeyCommitment(publicKeys)

	return &AccountInclusionValue{
		AccountIDPacked:     *accountIDPacked,
		AccountIDHash:       accountIDHash,
		AccountTreeRoot:     accountTreeRoot,
		AccountMerkleProofs: accountMerkleProofs,
		PublicKeys:          publicKeys,
		PublicKeyCommitment: publicKeyCommitment,
		IsValid:             true,
	}, nil
}
