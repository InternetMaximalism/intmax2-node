package types

import (
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"
)

type BlockWitnessFlatten struct {
	Block                   *block_post_service.PostedBlock     `json:"block"`
	Signature               SignatureContent                    `json:"signature"`
	PublicKeys              []intMaxTypes.Uint256               `json:"pubkeys"`
	PrevAccountTreeRoot     *intMaxTree.PoseidonHashOut         `json:"prevAccountTreeRoot"`
	PrevBlockTreeRoot       *intMaxTree.PoseidonHashOut         `json:"prevBlockTreeRoot"`
	AccountIdPacked         *AccountIdPacked                    `json:"accountIdPacked,omitempty"` // in account id case
	AccountMerkleProofs     []AccountMerkleProof                `json:"accountMerkleProofs"`       // in account id case
	AccountMembershipProofs []intMaxTree.IndexedMembershipProof `json:"accountMembershipProofs"`   // in pubkey case
}
