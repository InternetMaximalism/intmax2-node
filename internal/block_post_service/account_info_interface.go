package block_post_service

import (
	intMaxAcc "intmax2-node/internal/accounts"

	"github.com/holiman/uint256"
)

type AccountInfo interface {
	RegisterPublicKey(pk *intMaxAcc.PublicKey) (err error)
	PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	AccountBySenderAddress(senderAddress string) (accID *uint256.Int, err error)
}
