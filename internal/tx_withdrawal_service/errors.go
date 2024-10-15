package tx_withdrawal_service

import "errors"

var ErrFailedToGetBalance = errors.New("failed to get balance")

var ErrTokenNotFound = errors.New("token not found")

var ErrBlockNotFound = errors.New("block not found")

var ErrFailedToDecodeFromBase64 = errors.New("failed to decode from base64")

var ErrFailedToUnmarshal = errors.New("failed to unmarshal")
