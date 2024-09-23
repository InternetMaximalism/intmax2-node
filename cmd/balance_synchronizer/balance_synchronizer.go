package balance_synchronizer

import (
	"context"
	"fmt"
	"github.com/dimiro1/health"
	"github.com/spf13/cobra"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	intMaxAcc "intmax2-node/internal/accounts"
	"intmax2-node/internal/balance_prover_service"
	service "intmax2-node/internal/balance_synchronizer"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	mDBApp "intmax2-node/internal/mnemonic_wallet/models"
	"sync"
)

const (
	int5Key = 5
)

type Synchronizer struct {
	Context context.Context
	Cancel  context.CancelFunc
	WG      *sync.WaitGroup
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
	HC      *health.Handler
}

func NewSynchronizerCmd(s *Synchronizer) *cobra.Command {
	const (
		use   = "synchronizer"
		short = "run balance synchronizer command"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			s.Log.Debugf("Run Block Builder command\n")

			// err := s.SB.CheckScrollPrivateKey(s.Context)
			// if err != nil {
			// 	const msg = "check private key error occurred: %v"
			// 	s.Log.Fatalf(msg, err.Error())
			// }

			wg := sync.WaitGroup{}

			blockSynchronizer, err := block_synchronizer.NewBlockSynchronizer(
				s.Context, s.Config, s.Log,
			)
			if err != nil {
				const msg = "failed to get Block Builder IntMax Address: %+v"
				s.Log.Fatalf(msg, err.Error())
			}
			blockValidityService, err := block_validity_prover.NewBlockValidityService(s.Context, s.Config, s.Log, s.SB, s.DbApp)
			if err != nil {
				const msg = "failed to get Block Builder IntMax Address: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			/**
			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()

				timeout := 1 * time.Second
				ticker := time.NewTicker(timeout)
				blockNumber := uint32(1)
				for {
					select {
					case <-s.Context.Done():
						ticker.Stop()
						s.Log.Warnf("Received cancel signal from context, stopping...")
						return
					case <-ticker.C:
						fmt.Printf("===============blockNumber (balance synchronizer): %d\n", blockNumber)
						err = blockValidityService.SyncBlockProverWithBlockNumber(blockNumber)
						if err != nil {
							fmt.Printf("===============err (balance synchronizer): %v\n", err)
							if err.Error() == block_validity_prover.ErrNoValidityProofByBlockNumber.Error() {
								s.Log.Warnf("no last validity proof")
								time.Sleep(int5Key * time.Second)

								continue
							}

							if err.Error() == "block number is not equal to the last block number + 1" {
								s.Log.Warnf("block number is not equal to the last block number + 1")
								time.Sleep(int5Key * time.Second)

								continue
							}

							if strings.Contains(err.Error(), "block content by block number error") {
								s.Log.Warnf("block content by block number error")
								time.Sleep(int5Key * time.Second)

								continue
							}

							const msg = "failed to sync block prover: %+v"
							s.Log.Fatalf(msg, err.Error())
						}

						fmt.Printf("update blockNumber: %d\n", blockNumber)
						blockNumber++
					}
				}
			}()
			*/

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()

				var userWallet *mDBApp.Wallet
				userWallet, err = mnemonic_wallet.New().WalletFromPrivateKeyHex(s.Config.Wallet.PrivateKeyHex)
				if err != nil {
					const msg = "failed to get Block Builder IntMax Address: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				fmt.Printf("my Ethereum address: %s\n", userWallet.WalletAddress)
				fmt.Printf("my INTMAX address: %s\n", userWallet.IntMaxWalletAddress)

				// withdrawalAggregator, err := withdrawal_service.NewWithdrawalAggregatorService(
				// 	s.Context, s.Config, s.Log, s.DbApp, s.SB,
				// )
				// if err != nil {
				// 	const msg = "failed to create withdrawal aggregator service: %+v"
				// 	s.Log.Fatalf(msg, err.Error())
				// }
				// synchronizer := balance_synchronizer.NewSynchronizerDummy(s.Context, s.Config, s.Log, s.SB, s.DbApp)
				// synchronizer.TestE2E(blockValidityService, blockSynchronizer, blockBuilderWallet, withdrawalAggregator)

				// balanceProverService := balance_prover_service.NewBalanceProverService(s.Context, s.Config, s.Log, blockBuilderWallet)
				balanceProcessor := balance_prover_service.NewBalanceProcessor(
					s.Context, s.Config, s.Log,
				)

				var userPrivateKey *intMaxAcc.PrivateKey
				userPrivateKey, err = intMaxAcc.NewPrivateKeyFromString(userWallet.IntMaxPrivateKey)
				if err != nil {
					const msg = "failed to get IntMax Private Key: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				// userWalletState := service.NewUserWalletState(userPrivateKey)
				userWalletState, err := service.NewMockWallet(userPrivateKey)
				if err != nil {
					const msg = "failed to get Mock Wallet: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				syncBalanceProver := service.NewSyncBalanceProver(s.Context, s.Config, s.Log)

				balanceSynchronizer := service.NewSynchronizer(s.Context, s.Config, s.Log, s.SB, blockSynchronizer, blockValidityService, balanceProcessor, syncBalanceProver, userWalletState)
				err = balanceSynchronizer.Sync(userPrivateKey)
				if err != nil {
					const msg = "failed to sync: %+v"
					s.Log.Fatalf(msg, err.Error())
				}
			}()

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()
				if err = s.Init(); err != nil {
					const msg = "failed to start api: %+v"
					s.Log.Fatalf(msg, err.Error())
				}
			}()

			wg.Wait()
		},
	}
}

func (s *Synchronizer) Init() error {
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
