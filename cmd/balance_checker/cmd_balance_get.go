package balance_checker

import (
	"github.com/spf13/cobra"
)

func getBalanceCmd(b *Balance) *cobra.Command {
	const (
		use   = "get"
		short = "Manage of Get balance of specified INTMAX account"

		ethTokenType     = "eth"
		erc20TokenType   = "erc20"
		erc721TokenType  = "erc721"
		erc1155TokenType = "erc1155"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.AddCommand(getTokenBalanceCmd(b, ethTokenType))
	cmd.AddCommand(getTokenBalanceCmd(b, erc20TokenType))
	cmd.AddCommand(getTokenBalanceCmd(b, erc721TokenType))
	cmd.AddCommand(getTokenBalanceCmd(b, erc1155TokenType))

	return &cmd
}
