package server

import (
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=mock_block_synchronizer.go -package=server -source=block_synchronizer.go

type BlockSynchronizer interface {
	BackupTransaction(
		sender intMaxAcc.Address,
		encodedEncryptedTx string,
		signature string,
		blockNumber uint64,
	) error
	BackupTransfer(
		recipient intMaxAcc.Address,
		encodedEncryptedTransfer string,
		blockNumber uint64,
	) error
}
