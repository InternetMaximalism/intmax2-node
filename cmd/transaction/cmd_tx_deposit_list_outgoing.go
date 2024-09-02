package transaction

import "github.com/spf13/cobra"

func txDepositListOutgoingCmd(b *Transaction) *cobra.Command {
	const (
		use   = "outgoing"
		short = "Get deposit list (outgoing)"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {}

	return &cmd
}
