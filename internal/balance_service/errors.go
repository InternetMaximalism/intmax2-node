package balance_service

const (
	ErrTokenTypeRequired       = "Error: Token type is required"
	ErrETHBalanceCheckArgs     = "Error: ETH balance check doesn't require additional arguments"
	ErrERC20BalanceCheckArgs   = "Error: ERC20 balance check requires a token address"
	ErrERC721BalanceCheckArgs  = "Error: ERC721 balance check requires a token address and token ID"
	ErrERC1155BalanceCheckArgs = "Error: ERC1155 balance check requires a token address and token ID"
	ErrInvalidTokenType        = "Error: Invalid token type. Use 'eth', 'erc20', 'erc721', or 'erc1155'"
	ErrTokenNotFound           = "Error: Token not found on INTMAX network"
	ErrFailedToGetBalance      = "Error: Failed to get balance"
)
