package transaction

import (
	"fmt"
	"intmax2-node/pkg/utils"
	"os"

	"github.com/spf13/cobra"
)

func txDepositCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "deposit"
		short                  = "Send deposit transaction"
		amountKey              = "amount"
		emptyKey               = ""
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient INTMAX address. use as --recipient \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "private-key"
		userPrivateDescription = "specify user's Ethereum private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var amount string
	cmd.PersistentFlags().StringVar(&amount, amountKey, emptyKey, amountDescription)

	var recipientAddressStr string
	cmd.PersistentFlags().StringVar(&recipientAddressStr, recipientKey, emptyKey, recipientDescription)

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		err := newCommands().SendDepositTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, utils.RemoveZeroX(userEthPrivateKey))
		if err != nil {
			const msg = "Fatal: %v\n"
			fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
	}

	return &cmd
}
