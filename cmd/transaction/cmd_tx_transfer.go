package transaction

import (
	"github.com/spf13/cobra"
)

func txTransferCmd(b *Transaction) *cobra.Command {
	const (
		use   = "transfer"
		short = "Manage transfer transaction"

		ethTokenType     = "eth"
		erc20TokenType   = "erc20"
		erc721TokenType  = "erc721"
		erc1155TokenType = "erc1155"
	)

	transferCmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	transferCmd.AddCommand(txTransferTokenCmd(b, ethTokenType))
	transferCmd.AddCommand(txTransferTokenCmd(b, erc20TokenType))
	transferCmd.AddCommand(txTransferTokenCmd(b, erc721TokenType))
	transferCmd.AddCommand(txTransferTokenCmd(b, erc1155TokenType))

	return &transferCmd
}
