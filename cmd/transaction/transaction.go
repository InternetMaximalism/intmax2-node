package transaction

import (
	"context"
	"github.com/spf13/cobra"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
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
