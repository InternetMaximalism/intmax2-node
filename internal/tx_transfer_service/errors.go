package tx_transfer_service

import "errors"

var ErrRecoverWalletFromPrivateKey = errors.New("fail to recover INTMAX private key")

var ErrBlockNotFound = errors.New("block not found")
