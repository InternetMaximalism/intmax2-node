package deposit

import (
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"

	"github.com/spf13/cobra"
)

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

		err := utils.IsValidEthereumPrivateKey(d.Config.Blockchain.DepositAnalyzerPrivateKeyHex)
		if err != nil {
			const msg = "check deposit analyzer private key error occurred: %v"
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
