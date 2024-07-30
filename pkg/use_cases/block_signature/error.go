package block_signature

import "errors"

var ErrInvalidSignature = errors.New("fail to verify signature")

var ErrInvalidEnoughBalanceProof = errors.New("invalid enough balance proof")

// ErrSignTxTreeByAvailableFileFail error: failed to sign of tx tree by available file.
var ErrSignTxTreeByAvailableFileFail = errors.New("failed to sign of tx tree by available file")
