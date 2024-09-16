package balance_checker

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func getTokenBalanceCmd(b *Balance, token string) *cobra.Command {
	const (
		short = "Get balance by token %q of specified INTMAX account"

		emptyKey               = ""
		userPrivateKeyKey      = "private-key"
		userPrivateDescription = "specify user address. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   token,
		Short: fmt.Sprintf(short, token),
	}

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := b.SB.SetupEthereumNetworkChainID(b.Context)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}

		err = newCommands().GetBalance(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context,
			append([]string{token}, args...),
			userEthPrivateKey,
		)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Fatal: %v\n", err)
			os.Exit(1)
		}
	}

	return &cmd
}
