package server

import (
	intMaxAcc "intmax2-node/internal/accounts"
)

//go:generate mockgen -destination=mock_block_post_service.go -package=server -source=block_post_service.go

type BlockPostService interface {
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
