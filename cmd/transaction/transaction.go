package transaction

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

type Transaction struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	SB      ServiceBlockchain
	DbApp   block_validity_prover.SQLDriverApp
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
