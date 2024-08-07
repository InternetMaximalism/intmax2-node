package tx_withdrawal_service

import "errors"

const (
	ErrFailedToGetBalance = "failed to get balance"
)

var ErrTokenNotFound = errors.New("token not found")

var ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp = errors.New("failed to fetch balance by user address and token info with DBApp")
