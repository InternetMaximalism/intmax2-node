package messenger

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"

	"github.com/spf13/cobra"
)

type Messenger struct {
	Context context.Context
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
}

func NewMessengerCmd(m *Messenger) *cobra.Command {
	const (
		use   = "messenger"
		short = "Manage messenger"
	)

	messengerCmd := &cobra.Command{
		Use:   use,
		Short: short,
	}
	messengerCmd.AddCommand(relayerCmd(m))

	return messengerCmd
}

func relayerCmd(m *Messenger) *cobra.Command {
	const (
		use   = "relayer"
		short = "Run messenger relayer"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		l := m.Log.WithFields(logger.Fields{"module": use})

		err := m.SB.CheckEthereumPrivateKey(m.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = newCommands().MessengerRelayer(m.Context, m.Config, l, m.DbApp, m.SB).Do(m.Context)
		if err != nil {
			const msg = "failed to processing messenger relayer: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
