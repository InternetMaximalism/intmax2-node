package balance_synchronizer

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/intmax_block_content"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_block_validity_service_test.go -package=balance_synchronizer_test -source=block_validity_service.go

type BlockValidityService interface {
	BlockContentByTxRoot(txRoot common.Hash) (*intmax_block_content.PostedBlock, error)
	GetDepositInfoByHash(depositHash common.Hash) (depositInfo *block_validity_prover.DepositInfo, err error)
	FetchValidityProverInfo() (*block_validity_prover.ValidityProverInfo, error)
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*block_validity_prover.UpdateWitness, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
	ValidityPublicInputs(txRoot common.Hash) (validityPublicInputs *block_validity_prover.ValidityPublicInputs, senderLeaves []block_validity_prover.SenderLeaf, err error)
}
