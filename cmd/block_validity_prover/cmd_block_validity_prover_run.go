package block_validity_prover

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/docs/swagger"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	"intmax2-node/internal/l2_batch_index"
	"intmax2-node/internal/pb/gateway"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	"intmax2-node/internal/pb/listener"
	server "intmax2-node/pkg/grpc_server/block_validity_prover_server"
	"intmax2-node/third_party"
	"sync"
	"time"

	"github.com/dimiro1/health"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// const timeoutFailedToSyncBlockProver = 5

func blockValidityProverRun(s *Settings) *cobra.Command {
	const (
		use   = "run"
		short = "Run the Block Validity Prover"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			err := s.SB.SetupScrollNetworkChainID(s.Context)
			if err != nil {
				const msg = "init the scroll network by chain ID error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			wg := sync.WaitGroup{}

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()
				l2BI := l2_batch_index.New(s.Config, s.DbApp, s.SB)
				err = l2BI.Start(s.Context)
				if err != nil {
					const msg = "starting of the Batch Index Processing error occurred: %v"
					s.Log.Fatalf(msg, err.Error())
				}
			}()

			s.Log.Infof("Start Block Validity Prover")
			var blockValidityProver block_validity_prover.BlockValidityProver
			blockValidityProver, err = block_validity_prover.NewBlockValidityProver(
				s.Context, s.Config, s.Log, s.SB, s.DbApp,
			)
			if err != nil {
				const msg = "failed to start Block Validity Prover: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			blockValidityService, err := block_validity_prover.NewBlockValidityService(
				s.Context, s.Config, s.Log, s.SB, s.DbApp,
			)
			if err != nil {
				const msg = "failed to start Block Validity Service: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			blockNumber, err := blockValidityService.LatestSynchronizedBlockNumber()
			if err != nil {
				const msg = "failed to get the latest synchronized block number: %+v"
				s.Log.Fatalf(msg, err.Error())
			}
			blockNumber += 1
			fmt.Printf("blockNumber (server): %d\n", blockNumber)

			// wg.Add(1)
			// s.WG.Add(1)
			// go func() {
			// 	defer func() {
			// 		wg.Done()
			// 		s.WG.Done()
			// 	}()

			// 	timeout := 1 * time.Second
			// 	ticker := time.NewTicker(timeout)
			// 	for {
			// 		select {
			// 		case <-s.Context.Done():
			// 			ticker.Stop()
			// 			s.Log.Warnf("Received cancel signal from context, stopping...")
			// 			return
			// 		case <-ticker.C:
			// 			s.Log.Debugf("===============blockNumber: %d", blockNumber)
			// 			err = blockValidityService.SyncBlockProverWithBlockNumber(blockNumber)
			// 			if err != nil {
			// 				s.Log.Debugf("===============err: %s", err.Error())
			// 				if err.Error() == block_validity_prover.ErrNoValidityProofByBlockNumber.Error() {
			// 					s.Log.Warnf("no last validity proof")
			// 					time.Sleep(timeoutFailedToSyncBlockProver * time.Second)

			// 					continue
			// 				}

			// 				if err.Error() == "block number is not equal to the last block number + 1" {
			// 					s.Log.Warnf("block number is not equal to the last block number + 1")
			// 					time.Sleep(timeoutFailedToSyncBlockProver * time.Second)

			// 					continue
			// 				}

			// 				if strings.HasPrefix(err.Error(), "block content by block number error") {
			// 					s.Log.Warnf("block content by block number error")
			// 					time.Sleep(timeoutFailedToSyncBlockProver * time.Second)

			// 					continue
			// 				}

			// 				const msg = "failed to sync block prover: %+v"
			// 				s.Log.Fatalf(msg, err.Error())
			// 			}

			// 			s.Log.Debugf("update blockNumber: %d\n", blockNumber)
			// 			blockNumber++
			// 		}
			// 	}
			// }()

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()

				var nextSynchronizedDepositIndex uint32
				nextSynchronizedDepositIndex, err = blockValidityService.FetchNextDepositIndex()
				if err != nil {
					const msg = "failed to fetch last deposit index: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				var useTicker bool
				timeout := 5 * time.Second
				ticker := time.NewTicker(timeout)
				for {
					select {
					case <-s.Context.Done():
						ticker.Stop()
						return
					case <-ticker.C:
						if useTicker {
							continue
						}
						go func() {
							useTicker = true
							defer func() {
								useTicker = false
							}()

							s.Log.Debugf("balance validity ticker.C")
							err = blockValidityProver.SyncDepositedEvents()
							if err != nil {
								const msg = "failed to sync deposited events: %+v"
								s.Log.Fatalf(msg, err.Error())
							}

							err = blockValidityProver.SyncDepositTree(nil, nextSynchronizedDepositIndex)
							if err != nil {
								const msg = "failed to sync deposit tree: %+v"
								s.Log.Fatalf(msg, err.Error())
							}
						}()
					}
				}
			}()

			var bps block_synchronizer.BlockSynchronizer
			bps, err = block_synchronizer.NewBlockSynchronizer(s.Context, s.Config, s.Log)
			if err != nil {
				const msg = "failed to start Block Synchronizer: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
				}()

				tickerEventWatcher := time.NewTicker(s.Config.BlockValidityProver.TimeoutForEventWatcher)
				for {
					select {
					case <-s.Context.Done():
						tickerEventWatcher.Stop()
						return
					case <-tickerEventWatcher.C:
						fmt.Println("block content ticker.C")
						// sync block content
						var startBlock uint64
						startBlock, err := blockValidityProver.LastSeenBlockPostedEventBlockNumber()
						if err != nil {
							startBlock = s.Config.Blockchain.RollupContractDeployedBlockNumber
						}
						fmt.Printf("startBlock of LastSeenBlockPostedEventBlockNumber: %d\n", startBlock)

						var endBlock uint64
						endBlock, err = blockValidityProver.SyncBlockContent(bps, startBlock)
						if err != nil {
							panic(err)
						}
						fmt.Printf("endBlock of LastSeenBlockPostedEventBlockNumber: %d\n", endBlock)

						err = blockValidityProver.SetLastSeenBlockPostedEventBlockNumber(endBlock)
						if err != nil {
							var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
							panic(errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err))
						}

						fmt.Printf("Block %d is searched\n", endBlock)
					}
				}
			}()

			// wg.Add(1)
			// go func() {
			// 	defer func() {
			// 		wg.Done()
			// 	}()

			// 	err = p.SyncBlockValidityWitness()
			// 	if err != nil {
			// 		var ErrSyncBlockProverWithBlockNumberFail = errors.New("failed to sync block validity witness")
			// 		panic(errors.Join(ErrSyncBlockProverWithBlockNumberFail, err))
			// 	}
			// }()

			wg.Add(1)
			go func() {
				defer func() {
					wg.Done()
				}()

				err = blockValidityProver.SyncBlockValidityProof()
				if err != nil {
					var ErrSyncBlockValidityProofFail = errors.New("failed to sync block validity proof")
					panic(errors.Join(ErrSyncBlockValidityProofFail, err))
				}
			}()

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()
				if err = s.Init(blockValidityService); err != nil {
					const msg = "failed to start api: %+v"
					s.Log.Fatalf(msg, err.Error())
				}
			}()

			wg.Wait()
		},
	}
}

func (s *Settings) Init(bvs BlockValidityService) error {
	tm := time.Duration(s.Config.HTTP.Timeout) * time.Second

	var c *cors.Cors
	if s.Config.HTTP.CORSAllowAll {
		c = cors.AllowAll()
	} else {
		c = cors.New(cors.Options{
			AllowedOrigins:       s.Config.HTTP.CORS,
			AllowedMethods:       s.Config.HTTP.CORSAllowMethods,
			AllowedHeaders:       s.Config.HTTP.CORSAllowHeaders,
			ExposedHeaders:       s.Config.HTTP.CORSExposeHeaders,
			AllowCredentials:     s.Config.HTTP.CORSAllowCredentials,
			MaxAge:               s.Config.HTTP.CORSMaxAge,
			OptionsSuccessStatus: s.Config.HTTP.CORSStatusCode,
		})
	}

	srv := server.New(
		s.Log,
		s.Config,
		s.DbApp,
		server.NewCommands(),
		s.Config.HTTP.CookieForAuthUse,
		s.HC,
		s.SB,
		bvs,
	)
	ctx := context.WithValue(s.Context, consts.AppConfigs, s.Config)

	const (
		version   = "version"
		buildtime = "buildtime"
		app       = "app"
		appName   = " (node) "
		sqlDBApp  = "sql-db-app"
		checkSB   = "blockchain_service"
	)

	// run externals gRPC server listener
	grpcErr, gRPCServerStop := listener.Run(
		ctx,
		s.Log,
		appName,
		s.Config.GRPC.Addr(), // listen incoming host:port for gRPC server
		func(s grpc.ServiceRegistrar) {
			node.RegisterInfoServiceServer(s, srv)
			node.RegisterBlockValidityProverServiceServer(s, srv)
		},
	)

	// healthCheck
	s.HC.AddChecker(sqlDBApp, s.DbApp)
	s.HC.AddInfo(app, map[string]any{
		version:   buildvars.Version,
		buildtime: buildvars.BuildTime,
	})
	s.HC.AddChecker(checkSB, s.SB)

	// run web -> gRPC gateway
	gw, grpcGwErr := gateway.Run(
		ctx,
		&gateway.Params{
			Name:               appName,
			Logger:             s.Log,
			GatewayAddr:        s.Config.HTTP.Addr(), // listen incoming host:port for rest api
			DialAddr:           s.Config.GRPC.Addr(), // connect to gRPC server host:port
			HTTPTimeout:        tm,
			HealthCheckHandler: s.HC,
			Services: []gateway.RegisterServiceHandlerFunc{
				node.RegisterInfoServiceHandler,
				node.RegisterBlockValidityProverServiceHandler,
			},
			CorsHandler: c.Handler,
			Swagger: &gateway.Swagger{
				HostURL:            s.Config.Swagger.HostURL,
				BasePath:           s.Config.Swagger.BasePath,
				SwaggerPath:        configs.SwaggerBlockValidityProverPath,
				FsSwagger:          swagger.FsSwaggerBlockValidityProver,
				OpenAPIPath:        configs.SwaggerOpenAPIBlockValidityProverPath,
				FsOpenAPI:          third_party.OpenAPIBlockValidityProver,
				RegexpBuildVersion: s.Config.Swagger.RegexpBuildVersion,
				RegexpHostURL:      s.Config.Swagger.RegexpHostURL,
				RegexpBasePATH:     s.Config.Swagger.RegexpBasePATH,
			},
			Cookies: &http_response_modifier.Cookies{
				ForAuthUse:         s.Config.HTTP.CookieForAuthUse,
				Secure:             s.Config.HTTP.CookieSecure,
				Domain:             s.Config.HTTP.CookieDomain,
				SameSiteStrictMode: s.Config.HTTP.CookieSameSiteStrictMode,
			},
		},
	)

	const (
		start  = "%sapplication started (version: %s buildtime: %s)"
		finish = "%sapplication finished"
	)

	s.Log.Infof(start, appName, buildvars.Version, buildvars.BuildTime)
	defer s.Log.Infof(finish, appName)

	var err error
	select {
	case <-s.Context.Done():
	case err = <-grpcErr:
		const msg = "%sgRPC server error: %s"
		s.Log.Errorf(msg, appName, err)
	case err = <-grpcGwErr:
		const msg = "%sgRPC gateway error: %s"
		s.Log.Errorf(msg, appName, err)
	}

	if gw != nil {
		gw.SetStatus(health.Down)
	}

	gRPCServerStop()
	s.Cancel()

	return nil
}
