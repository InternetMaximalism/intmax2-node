package transaction

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"
	"os"
	"strings"

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
		use                         = "transfer"
		short                       = "Send transfer transaction"
		emptyKey                    = ""
		amountKey                   = "amount"
		amountDescription           = "specify amount without decimals. use as --amount \"10\""
		recipientKey                = "recipient"
		recipientDescription        = "specify recipient INTMAX address. use as --recipient \"0x0000000000000000000000000000000000000000000000000000000000000000\""
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

	var amount string
	cmd.PersistentFlags().StringVar(&amount, amountKey, emptyKey, amountDescription)

	var startBlockNumber string
	cmd.PersistentFlags().StringVar(&startBlockNumber, startBlockNumberKey, emptyKey, startBlockNumberDescription)

	var limit string
	cmd.PersistentFlags().StringVar(&limit, limitKey, emptyKey, limitDescription)

	var recipientAddressStr string
	cmd.PersistentFlags().StringVar(&recipientAddressStr, recipientKey, emptyKey, recipientDescription)

	var userEthPrivateKey string
	cmd.PersistentFlags().StringVar(&userEthPrivateKey, userPrivateKeyKey, emptyKey, userPrivateDescription)

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			const msg = "Fatal: unknown operation\n"
			fmt.Fprintf(os.Stderr, msg)
			os.Exit(1)
		}

		const (
			txTransferSenderListStr = "list"
			txTransferSenderInfoStr = "info"
		)

		switch strings.ToLower(args[0]) {
		case txTransferSenderListStr:
			resp, err := newCommands().SenderTransactionsList(b.Config, b.Log, b.SB).Do(b.Context, args, startBlockNumber, limit, utils.RemoveZeroX(userEthPrivateKey))
			if err != nil {
				const msg = "Fatal: %v\n"
				fmt.Fprintf(os.Stderr, msg, err)
				os.Exit(1)
			}
			_, _ = os.Stdout.WriteString(string(resp))
		case txTransferSenderInfoStr:
			const msg = "Fatal: unknown operation\n"
			fmt.Fprintf(os.Stderr, msg)
			os.Exit(1)
		default:
			err := newCommands().SendTransferTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, amount, recipientAddressStr, utils.RemoveZeroX(userEthPrivateKey))
			if err != nil {
				const msg = "Fatal: %v\n"
				fmt.Fprintf(os.Stderr, msg, err)
				os.Exit(1)
			}
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

func txWithdrawalCmd(b *Transaction) *cobra.Command {
	const (
		use                    = "withdrawal"
		short                  = "Send withdrawal transaction"
		amountKey              = "amount"
		emptyKey               = ""
		amountDescription      = "specify amount without decimals. use as --amount \"10\""
		recipientKey           = "recipient"
		recipientDescription   = "specify recipient Ethereum address. use as --recipient \"0x0000000000000000000000000000000000000000\""
		userPrivateKeyKey      = "private-key"
		userPrivateDescription = "specify user's private key. use as --private-key \"0x0000000000000000000000000000000000000000000000000000000000000000\""
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
		err := newCommands().SendWithdrawalTransaction(b.Config, b.Log, b.SB).Do(b.Context, args, recipientAddressStr, amount, utils.RemoveZeroX(userEthPrivateKey), resume)
		if err != nil {
			const msg = "Fatal: %v\n"
			fmt.Fprintf(os.Stderr, msg, err)
			os.Exit(1)
		}
	}

	return &cmd
}

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
