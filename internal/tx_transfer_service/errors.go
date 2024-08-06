package tx_transfer_service

import "errors"

const (
	ErrFailedToGetBalance = "failed to get balance"
)

// ErrTokenNotFound error: token not found.
var ErrTokenNotFound = errors.New("token not found")

// ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp error: failed to fetch balance by user address and token info with DBApp.
var ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp = errors.New("failed to fetch balance by user address and token info with DBApp")

// ErrTxTreeNotBuild error: the tx tree not build.
var ErrTxTreeNotBuild = errors.New("txHash: the tx tree not build.")

var ErrBlockNotFound = errors.New("block not found")
