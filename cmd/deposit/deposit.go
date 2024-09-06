package deposit

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

type Deposit struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
}

func NewDepositCmd(d *Deposit) *cobra.Command {
	const (
		use   = "deposit"
		short = "Manage deposit"
	)

	depositCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	depositCmd.AddCommand(analyzerCmd(d))

	return depositCmd
}
