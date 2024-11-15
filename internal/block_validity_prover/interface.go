package block_validity_prover

import (
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/intmax_block_content"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
)

type BlockValidityProver interface {
	SyncDepositedEvents() error
	SyncDepositTree(endBlock *uint64, depositIndex uint32) error
	SyncBlockContent() (lastEventSeenBlockNumber uint64, err error)
	SyncBlockValidityProof() error
	UpdateValidityWitness(
		blockContent *intMaxTypes.BlockContent,
		prevValidityWitness *ValidityWitness,
	) (*ValidityWitness, error)
	LastSeenBlockPostedEventBlockNumber() (uint64, error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
}

type BlockValidityService interface {
	BlockContentByTxRoot(txRoot common.Hash) (*intmax_block_content.PostedBlock, error)
	GetDepositInfoByHash(depositHash common.Hash) (depositInfo *DepositInfo, err error)
	GetDepositsInfoByHash(depositHash ...common.Hash) (map[uint32]*DepositInfo, error)
	FetchValidityProverInfo() (*ValidityProverInfo, error)
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error)
	DepositTreeProof(blockNumber uint32, depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
	ValidityPublicInputsByBlockNumber(
		blockNumber uint32,
	) (*ValidityPublicInputs, []SenderLeaf, error)
	ValidityPublicInputs(txRoot common.Hash) (validityPublicInputs *ValidityPublicInputs, senderLeaves []SenderLeaf, err error)
	ValidityProofByBlockNumber(blockNumber uint32) (*ValidityProof, error)
	AuxInfoListFromBlockContentByTxRoot(txRoot ...common.Hash) (map[common.Hash]*AuxInfo, error)
}
