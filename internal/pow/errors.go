package pow

import "errors"

// ErrMessageLengthInvalid error: invalid message length.
var ErrMessageLengthInvalid = errors.New("invalid message length")

// ErrPoWTrailingZerosTargetInvalid error: invalid trailing zeros target for compute PoW.
var ErrPoWTrailingZerosTargetInvalid = errors.New("invalid trailing zeros target for compute PoW")

// ErrPoWHashInvalid error: invalid hash for compute PoW.
var ErrPoWHashInvalid = errors.New("invalid hash for compute PoW")

// ErrConvertDigestFail error: failed to convert digest.
var ErrConvertDigestFail = errors.New("failed to convert digest")

// ErrAbsorbCurlP81Fail error: failed to absorb curl-p-81.
var ErrAbsorbCurlP81Fail = errors.New("failed to absorb curl-p-81")

// ErrDigestFail error: failed to get digest.
var ErrDigestFail = errors.New("failed to get digest")

// ErrDifficultyFail error: failed to get Difficulty.
var ErrDifficultyFail = errors.New("failed to get Difficulty")

// ErrScoreTargetInvalidFail error: invalid target score for compute PoW.
var ErrScoreTargetInvalidFail = errors.New("invalid target score for compute PoW")

// ErrSufficientTrailingZerosFail error: failed to sufficient trailing zeros.
var ErrSufficientTrailingZerosFail = errors.New("failed to sufficient trailing zeros")

// ErrConvertToIntFail error: failed to convert int.
var ErrConvertToIntFail = errors.New("failed to convert int")

// ErrStateToIntFail error: failed to state to int.
var ErrStateToIntFail = errors.New("failed to state to int")

// ErrCheckStateTritsFail error: failed to check state trits.
var ErrCheckStateTritsFail = errors.New("failed to check state trits")

// ErrMinePoWNonceFail error: failed to mine nonce with PoW workers.
var ErrMinePoWNonceFail = errors.New("failed to mine nonce with PoW workers")

// ErrScoreFail error: failed to score.
var ErrScoreFail = errors.New("failed to score")

// ErrPoWNonceInvalid error: PoW nonce invalid.
var ErrPoWNonceInvalid = errors.New("PoW nonce invalid")
