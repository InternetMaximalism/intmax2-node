package server

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/docs/swagger"
	"intmax2-node/internal/balance_prover_service"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/mnemonic_wallet"
	"intmax2-node/internal/pb/gateway"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	node "intmax2-node/internal/pb/gen/block_builder_service/node"
	"intmax2-node/third_party"
	"sync"
	"time"

	"github.com/dimiro1/health"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

type Synchronizer struct {
	Context          context.Context
	Cancel           context.CancelFunc
	WG               *sync.WaitGroup
	Config           *configs.Config
	Log              logger.Logger
	DbApp            SQLDriverApp
	BBR              BlockBuilderRegistryService
	SB               ServiceBlockchain
	NS               NetworkService
	HC               *health.Handler
	PoW              PoWNonce
	BlockPostService BlockPostService
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

			err := s.BlockPostService.Init(s.Context)
			if err != nil {
				const msg = "init the Block Validity Prover error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.SB.CheckScrollPrivateKey(s.Context)
			if err != nil {
				const msg = "check private key error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			err = s.NS.CheckNetwork(s.Context)
			if err != nil {
				const msg = "check network error occurred: %v"
				s.Log.Fatalf(msg, err.Error())
			}

			wg := sync.WaitGroup{}

			// s.Log.Infof("Start Block Validity Prover")
			// var blockValidityProver block_validity_prover.BlockValidityProver
			// blockValidityProver, err = block_validity_prover.NewBlockValidityProver(s.Context, s.Config, s.Log, s.SB, s.DbApp)
			// if err != nil {
			// 	const msg = "failed to start Block Validity Prover: %+v"
			// 	s.Log.Fatalf(msg, err.Error())
			// }
			syncValidityProver, err := balance_prover_service.NewSyncValidityProver(s.Context, s.Config, s.Log, s.SB, s.DbApp)
			if err != nil {
				s.Log.Fatalf("failed to create sync validity prover: %+v", err)
			}

			wg.Add(1)
			s.WG.Add(1)
			go func() {
				defer func() {
					wg.Done()
					s.WG.Done()
				}()

				blockBuilderWallet, err := mnemonic_wallet.New().WalletFromPrivateKeyHex(s.Config.Blockchain.BuilderPrivateKeyHex)
				if err != nil {
					const msg = "failed to get Block Builder IntMax Address: %+v"
					s.Log.Fatalf(msg, err.Error())
				}

				// withdrawalAggregator, err := withdrawal_service.NewWithdrawalAggregatorService(
				// 	s.Context, s.Config, s.Log, s.DbApp, s.SB,
				// )
				// if err != nil {
				// 	const msg = "failed to create withdrawal aggregator service: %+v"
				// 	s.Log.Fatalf(msg, err.Error())
				// }
				// synchronizer := balance_prover_service.NewSynchronizerDummy(s.Context, s.Config, s.Log, s.SB, s.DbApp)
				// synchronizer.TestE2E(syncValidityProver, blockBuilderWallet, withdrawalAggregator)

				synchronizer := balance_prover_service.NewSynchronizer(s.Context, s.Config, s.Log, s.SB, s.DbApp)
				err = synchronizer.Sync(syncValidityProver, blockBuilderWallet)
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

	// srv := server.New(
	// 	s.Log, s.Config, s.DbApp, server.NewCommands(), s.Config.HTTP.CookieForAuthUse, s.HC, s.PoW, s.Worker,
	// )
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

	// // run externals gRPC server listener
	// grpcErr, gRPCServerStop := listener.Run(
	// 	ctx,
	// 	s.Log,
	// 	appName,
	// 	s.Config.GRPC.Addr(), // listen incoming host:port for gRPC server
	// 	func(s grpc.ServiceRegistrar) {
	// 		node.RegisterInfoServiceServer(s, srv)
	// 		node.RegisterBlockBuilderServiceServer(s, srv)
	// 	},
	// )

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
	// case err = <-grpcErr:
	// 	const msg = "%sgRPC server error: %s"
	// 	s.Log.Errorf(msg, appName, err)
	case err = <-grpcGwErr:
		const msg = "%sgRPC gateway error: %s"
		s.Log.Errorf(msg, appName, err)
	}

	if gw != nil {
		gw.SetStatus(health.Down)
	}

	// gRPCServerStop()
	s.Cancel()

	return nil
}
