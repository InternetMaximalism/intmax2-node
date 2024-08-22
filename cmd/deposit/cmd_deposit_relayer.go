package deposit

import (
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"

	"github.com/spf13/cobra"
)

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

		err := utils.IsValidEthereumPrivateKey(d.Config.Blockchain.DepositRelayerPrivateKeyHex)
		if err != nil {
			const msg = "check deposit relayer private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = d.SB.SetupEthereumNetworkChainID(d.Context)
		if err != nil {
			const msg = "Fatal: %v"
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
