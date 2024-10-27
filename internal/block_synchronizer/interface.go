package block_synchronizer

import (
	"context"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/bindings"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type BlockSynchronizer interface {
	FetchLatestBlockNumber(ctx context.Context) (uint64, error)
	FetchNewPostedBlocks(startBlock uint64, endBlock *uint64) ([]*bindings.RollupBlockPosted, *big.Int, error)
	FetchScrollCalldataByHash(txHash common.Hash) ([]byte, error)
	BackupTransaction(
		sender intMaxAcc.Address,
		encodedEncryptedTxHash, encodedEncryptedTx string,
		senderLastBalanceProofBody, senderBalanceTransitionProofBody []byte,
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
