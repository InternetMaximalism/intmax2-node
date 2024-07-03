package withdrawal

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	service "intmax2-node/internal/withdrawal_service"

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
	withdrawalCmd.AddCommand(relayerCmd(w))

	return withdrawalCmd
}

func relayerCmd(w *Withdrawal) *cobra.Command {
	const (
		use   = "relayer"
		short = "Run withdrawal relayer service"
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

		service.WithdrawalRelayer(w.Context, w.Config, w.Log, w.SB)
	}

	return &cmd
}
