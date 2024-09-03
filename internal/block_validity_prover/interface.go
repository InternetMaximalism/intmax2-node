package block_validity_prover

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"intmax2-node/internal/block_post_service"
	intMaxTypes "intmax2-node/internal/types"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BlockValidityProver interface {
	SyncDepositedEvents() error
	SyncDepositTree(endBlock *uint64, depositIndex uint32) error
	// SyncBlockTree(_ BlockSynchronizer) (endBlock uint64, err error)
	SyncBlockTree(
		bps BlockSynchronizer,
		startBlock uint64,
	) (lastEventSeenBlockNumber uint64, err error)
	SyncBlockProverWithAuxInfo(
		blockContent *intMaxTypes.BlockContent,
		postedBlock *block_post_service.PostedBlock,
	) error
	SyncBlockProver(
		validityWitness *ValidityWitness,
	) error
	BlockBuilder() *mockBlockBuilder
	FetchLastDepositIndex() (uint32, error)
}

type BlockSynchronizer interface {
	FetchLatestBlockNumber(ctx context.Context) (uint64, error)
	FetchNewPostedBlocks(startBlock uint64, endBlock *uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
	RollupContractDeployedBlockNumber() uint64
	BackupTransaction(
		sender intMaxAcc.Address,
		encodedEncryptedTxHash, encodedEncryptedTx string,
		signature string,
		blockNumber uint64,
	) error
	BackupTransfer(
		recipient intMaxAcc.Address,
		encodedEncryptedTransferHash, encodedEncryptedTransfer string,
		blockNumber uint64,
	) error
	BackupWithdrawal(
		recipient common.Address,
		encodedEncryptedTransferHash, encodedEncryptedTransfer string,
		blockNumber uint64,
	) error
}
