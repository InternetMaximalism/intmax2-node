package block_validity_prover

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"sync"

	"github.com/dimiro1/health"

	"github.com/spf13/cobra"
)

const timeoutFailedToSyncBlockProver = 5

type Settings struct {
	Context             context.Context
	Cancel              context.CancelFunc
	WG                  *sync.WaitGroup
	Config              *configs.Config
	Log                 logger.Logger
	DbApp               SQLDriverApp
	SB                  ServiceBlockchain
	HC                  *health.Handler
	BlockPostService    BlockPostService
	DepositSynchronizer DepositSynchronizer
}

func NewCmd(s *Settings) *cobra.Command {
	const (
		use   = "block-validity-prover"
		short = "Manage the Block Validity Prover"
	)

	cmd := cobra.Command{
		Use:   use,
		Short: short,
	}

	cmd.AddCommand(blockValidityProverRun(s))

	return &cmd
}
