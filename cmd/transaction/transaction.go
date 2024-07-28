package transaction

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

type Transaction struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
}

func NewTransactionCmd(b *Transaction) *cobra.Command {
	const (
		use   = "tx"
		short = "Manage transaction"
	)

	depositCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	depositCmd.AddCommand(txTransferCmd(b))

	return depositCmd
}

func txTransferCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "transfer"
		short                  = "Send transfer transaction"
		amountKey              = "amount"
		emptyKey               = ""
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient address. use as --recipient \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "user-private"
		userPrivateDescription = "specify user address. use as --user-private \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var amount string
	cmd.PersistentFlags().StringVar(&amount, amountKey, emptyKey, amountDescription)

	var recipientAddressStr string
	cmd.PersistentFlags().StringVar(&recipientAddressStr, recipientKey, emptyKey, recipientDescription)

	var userPrivateKey string
	cmd.PersistentFlags().StringVar(&userPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := b.SB.CheckEthereumPrivateKey(b.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = b.DbApp.Exec(b.Context, nil, func(db interface{}, _ interface{}) (err error) {
			q := db.(SQLDriverApp)

			return newCommands().SendTransferTransaction(b.Config, b.Log, q, b.SB).Do(b.Context, args, amount, recipientAddressStr, removeZeroX(userPrivateKey))
		})
		if err != nil {
			const msg = "failed to get balance: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func removeZeroX(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	}
	return s
}
