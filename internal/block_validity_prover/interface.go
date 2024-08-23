package block_validity_prover

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BlockValidityProver interface {
	SyncDepositTree() error
	SyncBlockTree(_ BlockSynchronizer) error
}

type BlockSynchronizer interface {
	FetchLatestBlockNumber(ctx context.Context) (uint64, error)
	FetchNewPostedBlocks(startBlock uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
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
