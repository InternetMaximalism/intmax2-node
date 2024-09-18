package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txTransferTokenCmd(b *Transaction, token string) *cobra.Command {
	const (
		short = "Send transfer transaction by token %q"

		emptyKey               = ""
		amountKey              = "amount"
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient INTMAX address. use as --recipient \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "private-key"
		userPrivateDescription = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   token,
		Short: fmt.Sprintf(short, token),
	}

	var amount string
	cmd.PersistentFlags().StringVar(&amount, amountKey, emptyKey, amountDescription)

	var recipientAddressStr string
	cmd.PersistentFlags().StringVar(&recipientAddressStr, recipientKey, emptyKey, recipientDescription)

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := b.SB.SetupEthereumNetworkChainID(b.Context)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}

		err = newCommands().SendTransferTransaction(
			b.Config, b.Log, b.SB, b.DbApp,
		).Do(
			b.Context,
			append([]string{token}, args...),
			amount,
			recipientAddressStr,
			utils.RemoveZeroX(userEthPrivateKey),
		)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
	}

	return &cmd
}
