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
	messengerCmd.AddCommand(relayerMockCmd(m))
	messengerCmd.AddCommand(withdrawalRelayerCmd(m))
	messengerCmd.AddCommand(withdrawalRelayerMockCmd(m))

	return messengerCmd
}

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

		// TODO: Messenger Privet Key Check
		err := m.SB.CheckEthereumPrivateKey(m.Context)
		if err != nil {
			const msg = "check private key error occurred: %v"
			l.Fatalf(msg, err.Error())
		}

		err = newCommands().MessengerRelayerMock(m.Context, m.Config, l, m.DbApp, m.SB).Do(m.Context)
		if err != nil {
			const msg = "failed to processing messenger relayer mock: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func withdrawalRelayerCmd(w *Messenger) *cobra.Command {
	const (
		use   = "withdrawal-relayer"
		short = "Run messenger withdrawal relayer service"
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

		err = newCommands().MessengerWithdrawalRelayer(w.Context, w.Config, l, w.SB).Do(w.Context)
		if err != nil {
			const msg = "failed to processing messenger withdrawal relayer: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}

func withdrawalRelayerMockCmd(w *Messenger) *cobra.Command {
	const (
		use   = "withdrawal-relayer-mock"
		short = "Run messenger withdrawal relayer mock service"
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

		err = newCommands().MessengerWithdrawalRelayerMock(w.Context, w.Config, l, w.DbApp, w.SB).Do(w.Context)
		if err != nil {
			const msg = "failed to processing messenger withdrawal relayer mock: %v"
			l.Fatalf(msg, err.Error())
		}
	}

	return &cmd
}
