package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txListCmd(b *Transaction) *cobra.Command {
	const (
		use   = "list"
		short = "Get transactions list"

		emptyKey                    = ""
		userPrivateKeyKey           = "private-key"
		userPrivateDescription      = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		startBlockNumberKey         = "startBlockNumber"
		startBlockNumberDescription = "specify the start block number without decimals. use as --amount \"10\". only for `list` operation"
		limitKey                    = "limit"
		limitDescription            = "specify the limit without decimals. use as --amount \"10\". only for `list`"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var startBlockNumber string
	cmd.PersistentFlags().StringVar(&startBlockNumber, startBlockNumberKey, emptyKey, startBlockNumberDescription)

	var limit string
	cmd.PersistentFlags().StringVar(&limit, limitKey, emptyKey, limitDescription)

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		resp, err := newCommands().SenderTransactionsList(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context, args, startBlockNumber, limit, utils.RemoveZeroX(userEthPrivateKey),
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
