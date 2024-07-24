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
	depositCmd.AddCommand(relayerCmd(d))

	return depositCmd
}

func analyzerCmd(d *Deposit) *cobra.Command {
	const (
		use   = "analyzer"
		short = "Run deposit analyzer service"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := d.Log.WithFields(logger.Fields{"module": use})

		err := d.SB.CheckEthereumPrivateKey(d.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = d.DbApp.Exec(d.Context, nil, func(db interface{}, _ interface{}) (err error) {
			q := db.(SQLDriverApp)
			return newCommands().DepositAnalyzer(d.Config, l, q, d.SB).Do(d.Context)
		})
		if err != nil {
			const msg = "failed to processing deposit analyzer: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func relayerCmd(d *Deposit) *cobra.Command {
	const (
		use   = "relayer"
		short = "Run deposit relayer service"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := d.Log.WithFields(logger.Fields{"module": use})

		err := d.SB.CheckEthereumPrivateKey(d.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = d.DbApp.Exec(d.Context, nil, func(db interface{}, _ interface{}) (err error) {
			q := db.(SQLDriverApp)
			return newCommands().DepositRelayer(d.Config, l, q, d.SB).Do(d.Context)
		})
		if err != nil {
			const msg = "failed to processing deposit relayer: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
