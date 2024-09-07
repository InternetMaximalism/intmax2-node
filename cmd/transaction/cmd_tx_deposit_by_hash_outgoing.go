package transaction

import "github.com/spf13/cobra"

func txDepositByHashOutgoingCmd(b *Transaction) *cobra.Command {
	const (
		use   = "outgoing"
		short = "Get deposit by hash (outgoing); coming soon"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {}

	return &cmd
}
