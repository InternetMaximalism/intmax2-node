package balance_checker

import (
	"context"
	"github.com/spf13/cobra"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
)

type Balance struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	SB      ServiceBlockchain
}

func NewBalanceCmd(b *Balance) *cobra.Command {
	const (
		use   = "balance"
		short = "Manage balance"
	)

	balanceCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	balanceCmd.AddCommand(getBalanceCmd(b))

	return balanceCmd
}
