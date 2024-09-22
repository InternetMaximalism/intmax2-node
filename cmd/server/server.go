package server

import (
	"context"
	"errors"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/docs/swagger"
	"intmax2-node/internal/block_synchronizer"
	"intmax2-node/internal/block_validity_prover"
	errorsB "intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/gas_price_oracle"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/network_service"
	"intmax2-node/internal/pb/gateway"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/internal/pb/listener"
	"intmax2-node/pkg/grpc_server/server"
	"intmax2-node/third_party"
	"strings"
	"sync"
	"time"

	"github.com/dimiro1/health"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

const timeoutFailedToSyncBlockProver = 5

type Server struct {
	Context             context.Context
	Cancel              context.CancelFunc
	WG                  *sync.WaitGroup
	Config              *configs.Config
	Log                 logger.Logger
	DbApp               SQLDriverApp
	BBR                 BlockBuilderRegistryService
	SB                  ServiceBlockchain
	NS                  NetworkService
	HC                  *health.Handler
	PoW                 PoWNonce
	Worker              Worker
	DepositSynchronizer DepositSynchronizer
	GPOStorage          GPOStorage
	BlockPostService    BlockPostService
}

// nolint: gocyclo
func NewServerCmd(s *Server) *cobra.Command {
	const (
		use   = "run"
		short = "run command"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			s.Log.Debugf("Run Block Builder command\n")

			err := s.Worker.Init()
			if err != nil {
				const msg = "init the worker error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.BlockPostService.Init(s.Context)
			if err != nil {
				const msg = "init the Block Validity Prover error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.SB.CheckScrollPrivateKey(s.Context)
			if err != nil {
				const msg = "check private key error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.GPOStorage.Init(s.Context)
			if err != nil {
				const msg = "init the gas price oracle storage error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.GPOStorage.UpdValues(s.Context, gas_price_oracle.ScrollEthGPO)
			if err != nil {
				const msg = "failed to update values of the gas price oracle storage: %+v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.NS.CheckNetwork(s.Context)
			if err != nil {
				const msg = "check network error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			updBB := func() {
				errURL := s.BBR.UpdateBlockBuilder(s.Context, network_service.NodeExternalAddress.Address.Address())
				if errURL != nil {
					const msg = "update the Block Builder URL in blockchain error occurred: %v"
					if strings.Contains(errURL.Error(), errorsB.ErrInsufficientStakeAmountStr) {
						s.Log.Fatalf(msg, errorsB.ErrInsufficientStakeAmountStr)
					}
					s.Log.Fatalf(msg, errURL.Error())
				}
				const myAddrIs = "My address is %s"
				s.Log.Infof(myAddrIs, network_service.NodeExternalAddress.Address.Address())
			}
			s.Log.Infof("Start updBB")
			updBB()
			s.Log.Infof("Finish updBB")

			if network_service.NodeExternalAddress.Address.Type() ==
				network_service.NatDiscoverExternalAddressType {
				s.Log.Infof("NodeExternalAddress.Address.Type() == NatDiscoverExternalAddressType")
				go func() {
					ticker := time.NewTicker(configs.NatDiscoverReCheck)
					for {
						select {
						case <-s.Context.Done():
							ticker.Stop()
							return
						case <-ticker.C:
							var (
								errNatDAddr error
								newNatDAddr network_service.ExternalAddress
							)
							newNatDAddr, errNatDAddr = s.NS.NATDiscover(
								s.Context,
								network_service.NodeExternalAddress.Address.InternalPort(),
								network_service.NodeExternalAddress.Address,
							)
							if errNatDAddr != nil {
								const (
									maskNATDiscoverErr  = "%s"
									emptyNATDiscoverKey = ""
									int0NATDiscoverKey  = 0
								)
								s.Log.WithError(errNatDAddr).Errorf(maskNATDiscoverErr, network_service.ErrNATDiscoverFail)
								go func() {
									network_service.NodeExternalAddress.Lock()
									defer network_service.NodeExternalAddress.Unlock()
									network_service.NodeExternalAddress.Address = network_service.NewExternalAddress(
										nil, emptyNATDiscoverKey,
										int0NATDiscoverKey, int0NATDiscoverKey,
										network_service.NatDiscoverExternalAddressType,
										s.Config.Network.HTTPSUse || s.Config.HTTP.TLSUse,
									)

									s.Log.Infof("Start updBB (errNatDAddr != nil)")
									updBB()
									s.Log.Infof("End updBB (errNatDAddr != nil)")
								}()
								continue
							}
							go func() {
								network_service.NodeExternalAddress.Lock()
								defer network_service.NodeExternalAddress.Unlock()
								network_service.NodeExternalAddress.Address = newNatDAddr
								s.Log.Infof("Start updBB (errNatDAddr == nil)")
								updBB()
								s.Log.Infof("End updBB (errNatDAddr == nil)")
							}()
						}
					}
				}()
			}

			// TODO: add processing events of Rollup contract (example, BlockBuilderUpdated)

			wg := sync.WaitGroup{}

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()
				tickerCurrentFile := time.NewTicker(s.Config.Worker.TimeoutForCheckCurrentFile)
				defer func() {
					if tickerCurrentFile != nil {
						tickerCurrentFile.Stop()
					}
				}()
				tickerSignaturesAvailableFiles := time.NewTicker(s.Config.Worker.TimeoutForSignaturesAvailableFiles)
				defer func() {
					if tickerSignaturesAvailableFiles != nil {
						tickerSignaturesAvailableFiles.Stop()
					}
				}()
				if err = s.Worker.Start(s.Context, tickerCurrentFile, tickerSignaturesAvailableFiles); err != nil {
					const msg = "failed to start worker: %+v"
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
				tickerGasPriceOracle := time.NewTicker(s.Config.GasPriceOracle.Timeout)
				defer func() {
					if tickerGasPriceOracle != nil {
						tickerGasPriceOracle.Stop()
					}
				}()
				for {
					select {
					case <-s.Context.Done():
						return
					case <-tickerGasPriceOracle.C:
						err = s.GPOStorage.UpdValues(s.Context, gas_price_oracle.ScrollEthGPO)
						if err != nil {
							const msg = "failed to update values of the gas price oracle storage: %+v"
							s.Log.Fatalf(msg, err.Error())
						}
					}
				}
			}()

			// TODO: Occur error: Block range is too large
			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()
				tickerEventWatcher := time.NewTicker(s.Config.BlockPostService.TimeoutForPostingBlock)
				defer func() {
					if tickerEventWatcher != nil {
						tickerEventWatcher.Stop()
					}
				}()
				if err = s.BlockPostService.Start(s.Context, tickerEventWatcher); err != nil {
					const msg = "failed to start Block Validity Prover: %+v"
					s.Log.Fatalf(msg, err.Error())
				}
			}()

			s.Log.Infof("Start Block Validity Prover")
			// blockNumber := uint32(1)
			var blockValidityProver block_validity_prover.BlockValidityProver
			blockValidityProver, err = block_validity_prover.NewBlockValidityProver(s.Context, s.Config, s.Log, s.SB, s.DbApp)
			if err != nil {
				const msg = "failed to start Block Validity Prover: %+v"
				s.Log.Fatalf(msg, err.Error())
			}
			blockValidityService, err := block_validity_prover.NewBlockValidityService(s.Context, s.Config, s.Log, s.SB, s.DbApp)
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

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()

				timeout := 1 * time.Second
				ticker := time.NewTicker(timeout)
				for {
					select {
					case <-s.Context.Done():
						ticker.Stop()
						s.Log.Warnf("Received cancel signal from context, stopping...")
						return
					case <-ticker.C:
						fmt.Printf("===============blockNumber: %d\n", blockNumber)
						err = blockValidityService.SyncBlockProverWithBlockNumber(blockNumber)
						if err != nil {
							fmt.Printf("===============err: %v\n", err.Error())
							if err.Error() == block_validity_prover.ErrNoValidityProofByBlockNumber.Error() {
								s.Log.Warnf("no last validity proof")
								time.Sleep(timeoutFailedToSyncBlockProver * time.Second)

								continue
							}

							if err.Error() == "block number is not equal to the last block number + 1" {
								s.Log.Warnf("block number is not equal to the last block number + 1")
								time.Sleep(timeoutFailedToSyncBlockProver * time.Second)

								continue
							}

							if strings.HasPrefix(err.Error(), "block content by block number error") {
								s.Log.Warnf("block content by block number error")
								time.Sleep(timeoutFailedToSyncBlockProver * time.Second)

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

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()

				var blockSynchronizer block_synchronizer.BlockSynchronizer
				blockSynchronizer, err = block_synchronizer.NewBlockSynchronizer(s.Context, s.Config, s.Log)
				if err != nil {
					const msg = "failed to start Block Synchronizer: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				var latestSynchronizedDepositIndex uint32
				latestSynchronizedDepositIndex, err = blockValidityService.FetchLastDepositIndex()
				if err != nil {
					const msg = "failed to fetch last deposit index: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				timeout := 5 * time.Second
				ticker := time.NewTicker(timeout)
				for {
					select {
					case <-s.Context.Done():
						ticker.Stop()
						return
					case <-ticker.C:
						fmt.Println("balance validity ticker.C")
						err = blockValidityProver.SyncDepositedEvents()
						if err != nil {
							const msg = "failed to sync deposited events: %+v"
							s.Log.Fatalf(msg, err.Error())
						}

						err = blockValidityProver.SyncDepositTree(nil, latestSynchronizedDepositIndex)
						if err != nil {
							const msg = "failed to sync deposit tree: %+v"
							s.Log.Fatalf(msg, err.Error())
						}

						// sync block content
						var startBlock uint64
						startBlock, err = blockValidityService.LastSeenBlockPostedEventBlockNumber()
						if err != nil {
							startBlock = s.Config.Blockchain.RollupContractDeployedBlockNumber
							// var ErrLastSeenBlockPostedEventBlockNumberFail = errors.New("last seen block posted event block number fail")
							// panic(errors.Join(ErrLastSeenBlockPostedEventBlockNumberFail, err))
						}

						var endBlock uint64
						endBlock, err = blockValidityProver.SyncBlockTree(blockSynchronizer, startBlock)
						if err != nil {
							panic(err)
						}

						err = blockValidityService.SetLastSeenBlockPostedEventBlockNumber(endBlock)
						if err != nil {
							var ErrSetLastSeenBlockPostedEventBlockNumberFail = errors.New("set last seen block posted event block number fail")
							panic(errors.Join(ErrSetLastSeenBlockPostedEventBlockNumberFail, err))
						}

						fmt.Printf("Block %d is searched\n", endBlock)
					}
				}
			}()

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()
				tickerEventWatcher := time.NewTicker(s.Config.DepositSynchronizer.TimeoutForEventWatcher)
				defer func() {
					if tickerEventWatcher != nil {
						tickerEventWatcher.Stop()
					}
				}()
				if err = s.DepositSynchronizer.Start(tickerEventWatcher); err != nil {
					const msg = "failed to start Deposit Synchronizer: %+v"
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

func (s *Server) Init() error {
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
		s.PoW,
		s.Worker,
		s.SB,
		s.GPOStorage,
	)
	ctx := context.WithValue(s.Context, consts.AppConfigs, s.Config)

	const (
		version   = "version"
		buildtime = "buildtime"
		app       = "app"
		appName   = " (node) "
		sqlDBApp  = "sql-db-app"
		checkSB   = "blockchain_service"
		checkNS   = "network_service"
	)

	// run externals gRPC server listener
	grpcErr, gRPCServerStop := listener.Run(
		ctx,
		s.Log,
		appName,
		s.Config.GRPC.Addr(), // listen incoming host:port for gRPC server
		func(s grpc.ServiceRegistrar) {
			node.RegisterInfoServiceServer(s, srv)
			node.RegisterBlockBuilderServiceServer(s, srv)
		},
	)

	// healthCheck
	s.HC.AddChecker(sqlDBApp, s.DbApp)
	s.HC.AddInfo(app, map[string]any{
		version:   buildvars.Version,
		buildtime: buildvars.BuildTime,
	})
	s.HC.AddChecker(checkSB, s.SB)
	s.HC.AddChecker(checkNS, s.NS)

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
				node.RegisterBlockBuilderServiceHandler,
			},
			CorsHandler: c.Handler,
			Swagger: &gateway.Swagger{
				HostURL:            s.Config.Swagger.HostURL,
				BasePath:           s.Config.Swagger.BasePath,
				SwaggerPath:        configs.SwaggerBlockBuilderPath,
				FsSwagger:          swagger.FsSwaggerBlockBuilder,
				OpenAPIPath:        configs.SwaggerOpenAPIBlockBuilderPath,
				FsOpenAPI:          third_party.OpenAPIBlockBuilder,
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
