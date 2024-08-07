package tx_withdrawal_service

import "errors"

var ErrFailedToGetBalance = errors.New("failed to get balance")

var ErrTokenNotFound = errors.New("token not found")

var ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp = errors.New("failed to fetch balance by user address and token info with DBApp")

var ErrBlockNotFound = errors.New("block not found")
