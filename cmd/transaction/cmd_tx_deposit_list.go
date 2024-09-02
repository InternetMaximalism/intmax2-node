package transaction

import "github.com/spf13/cobra"

func txDepositListCmd(b *Transaction) *cobra.Command {
	const (
		use   = "list"
		short = "Get deposit list"
	)

	depositListCmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	depositListCmd.AddCommand(txDepositListIncomingCmd(b))
	depositListCmd.AddCommand(txDepositListOutgoingCmd(b))

	return &depositListCmd
}
