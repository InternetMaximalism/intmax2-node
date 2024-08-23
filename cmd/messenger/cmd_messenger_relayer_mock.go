package messenger

import (
	"intmax2-node/internal/logger"
	"intmax2-node/pkg/utils"

	"github.com/spf13/cobra"
)

func relayerMockCmd(m *Messenger) *cobra.Command {
	const (
		use   = "relayer-mock"
		short = "Run messenger relayer"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := m.Log.WithFields(logger.Fields{"module": use})

		err := utils.IsValidEthereumPrivateKey(m.Config.Blockchain.MessengerMockPrivateKeyHex)
		if err != nil {
			const msg = "check messenger private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = m.SB.SetupEthereumNetworkChainID(m.Context)
		if err != nil {
			const msg = "Fatal: %v"
			l.Fatalf(msg, err.Error())
		}

		err = m.SB.SetupScrollNetworkChainID(m.Context)
		if err != nil {
			const msg = "Fatal: %v"
			l.Fatalf(msg, err.Error())
		}

		err = newCommands().MessengerRelayerMock(m.Config, l, m.DbApp, m.SB).Do(m.Context)
		if err != nil {
			const msg = "failed to processing messenger relayer mock: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
