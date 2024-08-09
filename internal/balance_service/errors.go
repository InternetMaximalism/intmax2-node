package balance_service

import "errors"

const (
	ErrFailedToGetBalance                          = "failed to get balance on %s network"
	ErrFetchTokenByTokenAddressAndTokenIDWithDBApp = "Failed to fetch token by tokenAddress and tokenId with DBApp: %v\n"
)

var ErrInvalidTokenType = errors.New("invalid token type. Use 'eth', 'erc20', 'erc721', or 'erc1155'")
var ErrTokenNotFound = errors.New("token not found on INTMAX network")
var ErrFetchBalanceByUserAddressAndTokenInfoWithDBApp = errors.New("failed to fetch balance by user address and token info with DBApp")
var ErrInvalidPrivateKey = errors.New("invalid private key")
var ErrRecoverWalletFromPrivateKey = errors.New("fail to recover INTMAX private key")
var ErrFailedToGetTokenIndex = errors.New("failed to get token index")
var ErrFailedToGetETHBalance = errors.New("failed to get ETH token balance")
var ErrFailedToGetERC20Balance = errors.New("failed to get ERC20 token balance")
var ErrFailedToGetERC721Owner = errors.New("failed to get ERC721 token owner")
var ErrOwnerNotMatch = errors.New("owner of the token is not the same as the provided address")
var ErrFailedToGetERC1155Balance = errors.New("failed to get ERC1155 token balance")
var ErrDepositValidity = errors.New("failed to get deposit validity")
