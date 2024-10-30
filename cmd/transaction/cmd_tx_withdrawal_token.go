package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txWithdrawalTokenCmd(b *Transaction, token string) *cobra.Command {
	const (
		short = "Send withdrawal transaction by token %q"

		emptyKey               = ""
		amountKey              = "amount"
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient INTMAX address. use as --recipient \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "private-key"
		userPrivateDescription = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		resumeKey              = "resume"
		defaultResume          = false
		resumeDescription      = "resume withdrawal. use as --resume"
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

	var resume bool
	cmd.PersistentFlags().BoolVar(&resume, resumeKey, defaultResume, resumeDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := b.SB.SetupEthereumNetworkChainID(b.Context)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}

		err = newCommands().SendWithdrawalTransaction(
			b.Config, b.Log, b.SB,
		).Do(
			b.Context,
			append([]string{token}, args...),
			recipientAddressStr,
			amount,
			utils.RemoveZeroX(userEthPrivateKey),
			resume,
			b.DbApp, // TODO: Remove this dependency
		)
		if err != nil {
			const msg = "Fatal: %v\n"
			_, _ = fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
	}

	return &cmd
}
