package types

import (
	"intmax2-node/internal/block_post_service"
	intMaxGP "intmax2-node/internal/hash/goldenposeidon"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type CompressedBlockWitness struct {
	Block                              *block_post_service.PostedBlock      `json:"block"`
	Signature                          SignatureContent                     `json:"signature"`
	PublicKeys                         []intMaxTypes.Uint256                `json:"pubkeys"`
	PrevAccountTreeRoot                intMaxTree.PoseidonHashOut           `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot                  intMaxTree.PoseidonHashOut           `json:"prevBlockTreeRoot"`
	AccountIdPacked                    *AccountIdPacked                     `json:"accountIdPacked,omitempty"`                    // in account id case
	SignificantAccountMerkleProofs     *[]AccountMerkleProof                `json:"significantAccountMerkleProofs,omitempty"`     // in account id case
	SignificantAccountMembershipProofs *[]intMaxTree.IndexedMembershipProof `json:"significantAccountMembershipProofs,omitempty"` // in pubkey case
	CommonAccountMerkleProof           []*intMaxGP.PoseidonHashOut          `json:"commonAccountMerkleProof"`
}
