package balance_synchronizer

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/block_post_service"
	"intmax2-node/internal/block_validity_prover"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

//go:generate mockgen -destination=mock_interface_test.go -package=balance_synchronizer_test -source=interface.go

type BlockValidityService interface {
	BlockContentByTxRoot(txRoot common.Hash) (*block_post_service.PostedBlock, error)
	GetDepositInfoByHash(depositHash common.Hash) (depositInfo *block_validity_prover.DepositInfo, err error)
	FetchValidityProverInfo() (*block_validity_prover.ValidityProverInfo, error)
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber *uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*block_validity_prover.UpdateWitness, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.PoseidonMerkleProof, error)
	ValidityPublicInputs(txRoot common.Hash) (validityPublicInputs *block_validity_prover.ValidityPublicInputs, senderLeaves []block_validity_prover.SenderLeaf, err error)
}
