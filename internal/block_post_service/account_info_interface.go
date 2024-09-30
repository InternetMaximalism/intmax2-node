package block_post_service

import (
	intMaxAcc "intmax2-node/internal/accounts"

	"github.com/holiman/uint256"
)

type AccountInfo interface {
	// RegisterPublicKey(pk *intMaxAcc.PublicKey, lastSentBlockNumber uint32) (accID uint64, err error)
	PublicKeyByAccountID(blockNumber uint32, accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	AccountBySenderAddress(senderAddress string) (accID *uint256.Int, err error)
}
