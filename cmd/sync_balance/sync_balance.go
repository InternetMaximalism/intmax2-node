package sync_balance

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

type Balance struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
}

func NewBalanceCmd(b *Balance) *cobra.Command {
	const (
		use   = "balance"
		short = "Manage balance"
	)

	depositCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	depositCmd.AddCommand(syncBalanceCmd(b))

	return depositCmd
}

func syncBalanceCmd(b *Balance) *cobra.Command {
	const (
		use   = "sync"
		short = "Synchronize balance of specified INTMAX account"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := b.Log.WithFields(logger.Fields{"module": use})

		err := b.SB.CheckEthereumPrivateKey(b.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = b.DbApp.Exec(b.Context, nil, func(db interface{}, _ interface{}) (err error) {
			q := db.(SQLDriverApp)

			return newCommands().SyncBalance(b.Config, b.Log, q, b.SB).Do(b.Context, args)
		})
		if err != nil {
			const msg = "failed to processing synchronize balance: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
