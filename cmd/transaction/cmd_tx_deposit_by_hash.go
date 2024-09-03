package transaction

import "github.com/spf13/cobra"

func txDepositByHashCmd(b *Transaction) *cobra.Command {
	const (
		use   = "info"
		short = "Get deposit by hash"
	)

	depositByHashCmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	depositByHashCmd.AddCommand(txDepositByHashIncomingCmd(b))
	depositByHashCmd.AddCommand(txDepositByHashOutgoingCmd(b))

	return &depositByHashCmd
}
