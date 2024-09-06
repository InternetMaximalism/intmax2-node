package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txDepositByHashIncomingCmd(b *Transaction) *cobra.Command {
	const (
		use   = "incoming [DepositHash]"
		short = "Get deposit by hash (incoming)"

		emptyKey               = ""
		userPrivateKeyKey      = "private-key"
		userPrivateDescription = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			const msg = "Fatal: hash must setup as argument #1\n"
			_, _ = fmt.Fprintf(os.Stderr, msg)
			os.Exit(1)
		}

		resp, err := newCommands().ReceiverDepositByHashIncoming(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context, args, args[0], utils.RemoveZeroX(userEthPrivateKey),
		)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString(string(resp))
	}

	return &cmd
}
