package block_validity_prover

import (
	"context"
	"math/big"

	// intMaxAcc "intmax2-node/internal/accounts"

	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	intMaxTree "intmax2-node/internal/tree"

	"github.com/ethereum/go-ethereum/common"
)

type BlockValidityProver interface {
	SyncDepositedEvents() error
	SyncDepositTree(endBlock *uint64, depositIndex uint32) error
	SyncBlockTree(
		bps BlockSynchronizer,
		startBlock uint64,
	) (lastEventSeenBlockNumber uint64, err error)
	SyncBlockProverWithBlockNumber(blockNumber uint32) error
	SyncBlockProver() error
	PostBlock(isRegistrationBlock bool, txs []*MockTxRequest) (*ValidityWitness, error)
}

type BlockValidityService interface {
	FetchLastDepositIndex() (uint32, error)
	LastSeenBlockPostedEventBlockNumber() (uint64, error)
	SetLastSeenBlockPostedEventBlockNumber(blockNumber uint64) error
	LatestIntMaxBlockNumber() uint32
	BlockContentByTxRoot(txRoot string) (*block_post_service.PostedBlock, error)
	GetDepositLeafAndIndexByHash(depositHash common.Hash) (depositLeafWithId *DepositLeafWithId, depositIndex *uint32, err error)
	BlockNumberByDepositIndex(depositIndex uint32) (uint32, error)
	LatestSynchronizedBlockNumber() (uint32, error)
	IsSynchronizedDepositIndex(depositIndex uint32) (bool, error)
	FetchUpdateWitness(publicKey *intMaxAcc.PublicKey, currentBlockNumber uint32, targetBlockNumber uint32, isPrevAccountTree bool) (*UpdateWitness, error)
	DepositTreeProof(depositIndex uint32) (*intMaxTree.KeccakMerkleProof, common.Hash, error)
	BlockTreeProof(rootBlockNumber uint32, leafBlockNumber uint32) (*intMaxTree.MerkleProof, error)
	RollupContractDeployedBlockNumber() uint64
	PostBlock(isRegistrationBlock bool, txs []*MockTxRequest) (*ValidityWitness, error) // XXX
}

type BlockSynchronizer interface {
	FetchLatestBlockNumber(ctx context.Context) (uint64, error)
	FetchNewPostedBlocks(startBlock uint64, endBlock *uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
}
