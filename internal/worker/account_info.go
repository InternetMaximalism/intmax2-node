package worker

import (
	intMaxAcc "intmax2-node/internal/accounts"

	"github.com/holiman/uint256"
)

//go:generate mockgen -destination=mock_account_info_test.go -package=worker_test -source=account_info.go

type AccountInfo interface {
	RegisterPublicKey(pk *intMaxAcc.PublicKey) (err error)
	PublicKeyByAccountID(accountID uint64) (pk *intMaxAcc.PublicKey, err error)
	AccountBySenderAddress(senderAddress string) (accID *uint256.Int, err error)
}
