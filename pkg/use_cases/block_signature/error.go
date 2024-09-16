package block_signature

import "errors"

// ErrSignTxTreeByAvailableFileFail error: failed to sign of tx tree by available file.
var ErrSignTxTreeByAvailableFileFail = errors.New("failed to sign of tx tree by available file")

var ErrDecodeSenderLastBalanceProofBodyFail = errors.New("failed to decode sender last balance proof body")

var ErrDecodeSenderTransitionProofBodyFail = errors.New("failed to decode sender transition proof body")

var ErrBackupTransferFail = errors.New("failed to backup transfer")
