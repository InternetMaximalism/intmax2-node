package transaction

import (
	"github.com/spf13/cobra"
)

func txDepositCmd(b *Transaction) *cobra.Command {
	const (
		use   = "deposit"
		short = "Send deposit transaction"

		ethTokenType     = "eth"
		erc20TokenType   = "erc20"
		erc721TokenType  = "erc721"
		erc1155TokenType = "erc1155"
	)

	depositCmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	depositCmd.AddCommand(txDepositTokenCmd(b, ethTokenType))
	depositCmd.AddCommand(txDepositTokenCmd(b, erc20TokenType))
	depositCmd.AddCommand(txDepositTokenCmd(b, erc721TokenType))
	depositCmd.AddCommand(txDepositTokenCmd(b, erc1155TokenType))
	depositCmd.AddCommand(txDepositListCmd(b))

	return &depositCmd
}
