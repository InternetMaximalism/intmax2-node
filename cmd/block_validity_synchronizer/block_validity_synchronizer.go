package block_validity_synchronizer

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"sync"

	"github.com/dimiro1/health"
	"github.com/spf13/cobra"
)

// const (
// 	int5Key = 5
// )

type BlockValiditySynchronizer struct {
	Context context.Context
	Cancel  context.CancelFunc
	WG      *sync.WaitGroup
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
	HC      *health.Handler
}

func NewBlockValiditySynchronizerCmd(s *BlockValiditySynchronizer) *cobra.Command {
	const (
		use   = "block-synchronizer"
		short = "run block synchronizer command"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			s.Log.Infof("Start Block Validity Prover")
			var blockValidityProver block_validity_prover.BlockValidityProver
			blockValidityProver, err := block_validity_prover.NewBlockValidityProver(s.Context, s.Config, s.Log, s.SB, s.DbApp)
			if err != nil {
				const msg = "failed to start Block Validity Prover: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			var blockSynchronizer block_synchronizer.BlockSynchronizer
			blockSynchronizer, err = block_synchronizer.NewBlockSynchronizer(s.Context, s.Config, s.Log)
			if err != nil {
				const msg = "failed to start Block Validity Synchronizer: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			if len(args) == 0 {
				wg := sync.WaitGroup{}
				if err = blockValidityProver.SyncBlockTree(blockSynchronizer, &wg); err != nil {
					panic(err)
				}
				wg.Wait()
				return
			}

			step := args[0]
			if err = blockValidityProver.SyncBlockTreeStep(blockSynchronizer, step); err != nil {
				panic(err)
			}
		},
	}
}

func (s *BlockValiditySynchronizer) Init() error {
	const (
		version   = "version"
		buildtime = "buildtime"
		app       = "app"
		appName   = " (node) "
		sqlDBApp  = "sql-db-app"
		checkSB   = "blockchain_service"
		checkNS   = "network_service"
	)

	// healthCheck
	s.HC.AddChecker(sqlDBApp, s.DbApp)
	s.HC.AddInfo(app, map[string]any{
		version:   buildvars.Version,
		buildtime: buildvars.BuildTime,
	})
	s.HC.AddChecker(checkSB, s.SB)

	const (
		start  = "%sapplication started (version: %s buildtime: %s)"
		finish = "%sapplication finished"
	)

	s.Log.Infof(start, appName, buildvars.Version, buildvars.BuildTime)
	defer s.Log.Infof(finish, appName)

	<-s.Context.Done()

	s.Cancel()

	return nil
}
