package server

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/docs/swagger"
	"intmax2-node/internal/blockchain/errors"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/network_service"
	"intmax2-node/internal/pb/gateway"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	"intmax2-node/internal/pb/gen/service/node"
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
}

func NewServerCmd(s *Server) *cobra.Command {
	const (
		use   = "run"
		short = "run command"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			err := s.Worker.Init()
			if err != nil {
				const msg = "init the worker error occurred: %v"
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

			updBB := func() {
				errURL := s.BBR.UpdateBlockBuilder(s.Context, network_service.NodeExternalAddress.Address.Address())
				if errURL != nil {
					const msg = "update the Block Builder URL in blockchain error occurred: %v"
					if strings.Contains(errURL.Error(), errors.ErrInsufficientStakeAmountStr) {
						s.Log.Fatalf(msg, errors.ErrInsufficientStakeAmountStr)
					}
					s.Log.Fatalf(msg, errURL.Error())
				}
				const myAddrIs = "My address is"
				fmt.Println(myAddrIs, network_service.NodeExternalAddress.Address.Address())
			}
			updBB()

			if network_service.NodeExternalAddress.Address.Type() ==
				network_service.NatDiscoverExternalAddressType {
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
									updBB()
								}()
								continue
							}
							go func() {
								network_service.NodeExternalAddress.Lock()
								defer network_service.NodeExternalAddress.Unlock()
								network_service.NodeExternalAddress.Address = newNatDAddr
								updBB()
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
				tickerEventWatcher := time.NewTicker(s.Config.DepositSynchronizer.TimeoutForEventWatcher)
				defer func() {
					if tickerEventWatcher != nil {
						tickerEventWatcher.Stop()
					}
				}()
				if err = s.DepositSynchronizer.Start(s.Context, tickerEventWatcher); err != nil {
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
		s.Log, s.Config, s.DbApp, server.NewCommands(), s.Config.HTTP.CookieForAuthUse, s.HC, s.PoW, s.Worker,
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
				SwaggerPath:        configs.SwaggerPath,
				FsSwagger:          swagger.FsSwagger,
				OpenAPIPath:        configs.SwaggerOpenAPIPath,
				FsOpenAPI:          third_party.OpenAPI,
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
