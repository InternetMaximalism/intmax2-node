package block_validity_prover

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_service.go -package=block_validity_prover -source=block_validity_service.go

type BlockValidityService interface {
	LatestDepositTreeProofByBlockNumber(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
	FetchValidityProverInfo() (*block_validity_prover.ValidityProverInfo, error)
	FetchUpdateWitness(
		publicKey *intMaxAcc.PublicKey,
		currentBlockNumber uint32,
		targetBlockNumber uint32,
		isPrevAccountTree bool,
	) (*block_validity_prover.UpdateWitness, error)
	ValidityPublicInputsByBlockNumber(
		blockNumber uint32,
	) (
		validityPublicInputs *block_validity_prover.ValidityPublicInputs,
		senderLeaves []block_validity_prover.SenderLeaf,
		err error,
	)
	ValidityProofByBlockNumber(blockNumber uint32) (*block_validity_prover.ValidityProof, error)
}
