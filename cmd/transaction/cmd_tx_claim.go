package transaction

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func txClaimCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "claim"
		short                  = "Send claim transaction"
		userEthPrivateKeyKey   = "private-key"
		emptyKey               = ""
		userPrivateDescription = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userEthPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := newCommands().SendClaimWithdrawals(b.Config, b.Log, b.SB).Do(b.Context, args, userEthPrivateKey)
		if err != nil {
			const msg = "Fatal: %v\n"
			fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
	}

	return &cmd
}
