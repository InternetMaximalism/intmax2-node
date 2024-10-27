package block_validity_prover

import (
	"context"
	"math/big"
	"sync"

	// intMaxAcc "intmax2-node/internal/accounts"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"
	intMaxTypes "intmax2-node/internal/types"

	"github.com/ethereum/go-ethereum/common"
)

type BlockValidityProver interface {
	SyncDepositedEvents() error
	SyncDepositTree(endBlock *uint64, depositIndex uint32) error
	SyncBlockTree(
		bps BlockSynchronizer,
		wg *sync.WaitGroup,
	) (err error)
	SyncBlockTreeStep(
		bps BlockSynchronizer,
		step string,
	) error
	UpdateValidityWitness(
		blockContent *intMaxTypes.BlockContent,
		prevValidityWitness *ValidityWitness,
	) (*ValidityWitness, error)
	LastSeenBlockPostedEventBlockNumber() (uint64, error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
}

type BlockValidityService interface {
	BlockContentByTxRoot(txRoot common.Hash) (*block_post_service.PostedBlock, error)
	GetDepositInfoByHash(depositHash common.Hash) (depositInfo *DepositInfo, err error)
	FetchValidityProverInfo() (*ValidityProverInfo, error)
	// Returns an update witness for a given public key and block numbers.
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber *uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error)
	DepositTreeProof(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(
		rootBlockNumber, leafBlockNumber uint32,
	) (
		*intMaxTree.PoseidonMerkleProof,
		*intMaxTree.PoseidonHashOut,
		error,
	)
	// ValidityWitness(txRoot string) (*ValidityWitness, error)
	ValidityPublicInputs(txRoot common.Hash) (validityPublicInputs *ValidityPublicInputs, senderLeaves []SenderLeaf, err error)
}

type BlockSynchronizer interface {
	FetchLatestBlockNumber(ctx context.Context) (uint64, error)
	FetchNewPostedBlocks(startBlock uint64, endBlock *uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
}
