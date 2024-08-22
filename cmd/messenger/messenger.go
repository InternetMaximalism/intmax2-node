package messenger

import (
	"context"
	"github.com/spf13/cobra"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
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
