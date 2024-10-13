package transaction

import (
	"github.com/spf13/cobra"
)

func txWithdrawalCmd(b *Transaction) *cobra.Command {
	const (
		use   = "withdrawal"
		short = "Manage withdrawal transaction"

		ethTokenType     = "eth"
		erc20TokenType   = "erc20"
		erc721TokenType  = "erc721"
		erc1155TokenType = "erc1155"
	)

	withdrawalCmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	withdrawalCmd.AddCommand(txWithdrawalTransfersListCmd(b))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, ethTokenType))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, erc20TokenType))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, erc721TokenType))
	withdrawalCmd.AddCommand(txWithdrawalTokenCmd(b, erc1155TokenType))

	return &withdrawalCmd
}
