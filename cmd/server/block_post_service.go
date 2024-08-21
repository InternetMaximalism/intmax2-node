package server

import (
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=mock_block_post_service.go -package=server -source=block_post_service.go

type BlockPostService interface {
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
}
