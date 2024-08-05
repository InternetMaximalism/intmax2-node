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
	depositCmd.AddCommand(txDepositCmd(b))
	depositCmd.AddCommand(txWithdrawalCmd(b))
	depositCmd.AddCommand(txClaimCmd(b))

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
		userPrivateDescription = "specify user Ethereum address. use as --user-private \"0x0000000000000000000000000000000000000000000000000000000000000000\""
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
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := b.SB.CheckEthereumPrivateKey(b.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = newCommands().SendTransferTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, amount, recipientAddressStr, removeZeroX(userEthPrivateKey))
		if err != nil {
			const msg = "failed to transfer transaction: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func txDepositCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "deposit"
		short                  = "Send deposit transaction"
		amountKey              = "amount"
		emptyKey               = ""
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient INTMAX address. use as --recipient \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "user-private"
		userPrivateDescription = "specify user's Ethereum address. use as --user-private \"0x0000000000000000000000000000000000000000000000000000000000000000\""
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
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := b.SB.CheckEthereumPrivateKey(b.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = newCommands().SendDepositTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, removeZeroX(userEthPrivateKey))
		if err != nil {
			const msg = "failed to deposit transaction: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func txWithdrawalCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "withdrawal"
		short                  = "Send withdraw transaction"
		amountKey              = "amount"
		emptyKey               = ""
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient Ethereum address. use as --recipient \"0x0000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "user-private"
		userPrivateDescription = "specify user address. use as --user-private \"0x0000000000000000000000000000000000000000000000000000000000000000\""
		resumeKey              = "resume"
		defaultResume          = false
		resumeDescription      = "resume withdrawal. use as --resume"
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

	var resume bool
	cmd.PersistentFlags().BoolVar(&resume, resumeKey, defaultResume, resumeDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := newCommands().SendWithdrawalTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, removeZeroX(userEthPrivateKey), resume)
		if err != nil {
			const msg = "failed to get balance: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func txClaimCmd(b *Transaction) *cobra.Command {
	const (
		use                  = "claim"
		short                = "Send claim transaction"
		recipientKey         = "recipient"
		emptyKey             = ""
		recipientDescription = "specify recipient Ethereum address. use as --recipient \"0x0000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var recipientEthAddress string
	cmd.PersistentFlags().StringVar(&recipientEthAddress, recipientKey, emptyKey, recipientDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := newCommands().SendClaimWithdrawals(b.Config, b.Log, b.SB).Do(b.Context, args, recipientEthAddress)
		if err != nil {
			const msg = "failed to claim withdrawals: %v"
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
