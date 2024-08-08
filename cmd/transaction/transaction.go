package transaction

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"

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

	transactionCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	transactionCmd.AddCommand(txTransferCmd(b))
	transactionCmd.AddCommand(txDepositCmd(b))
	transactionCmd.AddCommand(txWithdrawalCmd(b))
	transactionCmd.AddCommand(txClaimCmd(b))

	return transactionCmd
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

		err := newCommands().SendTransferTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, amount, recipientAddressStr, utils.RemoveZeroX(userEthPrivateKey))
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

		err := newCommands().SendDepositTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, utils.RemoveZeroX(userEthPrivateKey))
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

		err := newCommands().SendWithdrawalTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, utils.RemoveZeroX(userEthPrivateKey), resume)
		if err != nil {
			const msg = "failed to get balance: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func txClaimCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "claim"
		short                  = "Send claim transaction"
		userEthPrivateKeyKey   = "user-private"
		emptyKey               = ""
		userPrivateDescription = "specify user's Ethereum address. use as --user-private \"0x0000000000000000000000000000000000000000000000000000000000000000\""
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userEthPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := newCommands().SendClaimWithdrawals(b.Config, b.Log, b.SB).Do(b.Context, args, userEthPrivateKey)
		if err != nil {
			const msg = "failed to claim withdrawals: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
