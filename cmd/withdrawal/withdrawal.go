package withdrawal

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

type Withdrawal struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
}

func NewWithdrawCmd(w *Withdrawal) *cobra.Command {
	const (
		use   = "withdrawal"
		short = "Manage withdrawal"
	)

	withdrawalCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	withdrawalCmd.AddCommand(aggregatorCmd(w))

	return withdrawalCmd
}

func aggregatorCmd(w *Withdrawal) *cobra.Command {
	const (
		use   = "aggregator"
		short = "Run withdrawal aggregator service"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := w.Log.WithFields(logger.Fields{"module": use})

		err := w.SB.CheckEthereumPrivateKey(w.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}
		err = w.DbApp.Exec(w.Context, nil, func(db interface{}, _ interface{}) (err error) {
			q := db.(SQLDriverApp)
			return newCommands().WithdrawalAggregator(w.Context, w.Config, l, q, w.SB).Do(w.Context)
		})
		if err != nil {
			const msg = "failed to processing withdrawal aggregator: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
