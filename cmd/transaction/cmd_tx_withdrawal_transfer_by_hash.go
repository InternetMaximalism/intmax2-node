package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func txWithdrawalTransferInfoByHashCmd(b *Transaction) *cobra.Command {
	const (
		use   = "info [TransferHash]"
		short = "Get withdrawal transfer by hash"

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

		resp, err := newCommands().RecipientWithdrawalTransferByHash(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context, args, args[0], utils.RemoveZeroX(userEthPrivateKey),
		)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				const msg = "%v\n"
				_, _ = fmt.Fprintf(os.Stderr, msg, err)
			} else {
				const msg = "Fatal: %v\n"
				_, _ = fmt.Fprintf(os.Stderr, msg, err)
			}
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString(string(resp))
	}

	return &cmd
}
