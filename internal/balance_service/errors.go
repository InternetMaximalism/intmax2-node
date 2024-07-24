package balance_service

const (
	//nolint:gosec
	ErrTokenTypeRequired                           = "token type is required"
	ErrETHBalanceCheckArgs                         = "ETH balance check doesn't require additional arguments"
	ErrERC20BalanceCheckArgs                       = "ERC20 balance check requires a token address"
	ErrERC721BalanceCheckArgs                      = "ERC721 balance check requires a token address and token ID"
	ErrERC1155BalanceCheckArgs                     = "ERC1155 balance check requires a token address and token ID"
	ErrInvalidTokenType                            = "invalid token type. Use 'eth', 'erc20', 'erc721', or 'erc1155'"
	ErrTokenNotFound                               = "token not found on INTMAX network"
	ErrFailedToGetBalance                          = "failed to get balance"
	ErrFetchTokenByTokenAddressAndTokenIDWithDBApp = "Failed to fetch token by tokenAddress and tokenId with DBApp: %v\n"
)
