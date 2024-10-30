package store_vault_server

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/docs/swagger"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gateway"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	node "intmax2-node/internal/pb/gen/store_vault_service/node"
	"intmax2-node/internal/pb/listener"
	server "intmax2-node/pkg/grpc_server/store_vault_server"
	"intmax2-node/third_party"
	"sync"
	"time"

	"github.com/dimiro1/health"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

type StoreVaultServer struct {
	Context context.Context
	Cancel  context.CancelFunc
	WG      *sync.WaitGroup
	Config  *configs.Config
	Log     logger.Logger
	DbApp   SQLDriverApp
	SB      ServiceBlockchain
	HC      *health.Handler
}

func NewServerCmd(s *StoreVaultServer) *cobra.Command {
	const (
		use   = "store-vault-server run"
		short = "run store vault server command"
	)
	return &cobra.Command{
		Use:   use,
		Short: short,
		Run: func(cmd *cobra.Command, args []string) {
			if err := s.Init(); err != nil {
				const msg = "failed to start api: %+v"
				s.Log.Fatalf(msg, err.Error())
			}
		},
	}
}

// TODO: Common parts should be shared between the server side and the client side.
func (s *StoreVaultServer) Init() error {
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
		s.Log, s.Config, s.DbApp, server.NewCommands(), s.SB, s.Config.HTTP.CookieForAuthUse, s.HC,
	)
	ctx := context.WithValue(s.Context, consts.AppConfigs, s.Config)

	const (
		version   = "version"
		buildtime = "buildtime"
		app       = "app"
		appName   = " (node) "
		sqlDBApp  = "sql-db-app"
	)

	// run externals gRPC server listener
	grpcErr, gRPCServerStop := listener.Run(
		ctx,
		s.Log,
		appName,
		s.Config.GRPC.Addr(), // listen incoming host:port for gRPC server
		func(s grpc.ServiceRegistrar) {
			node.RegisterInfoServiceServer(s, srv)
			node.RegisterStoreVaultServiceServer(s, srv)
		},
	)

	// healthCheck
	s.HC.AddChecker(sqlDBApp, s.DbApp)
	s.HC.AddInfo(app, map[string]any{
		version:   buildvars.Version,
		buildtime: buildvars.BuildTime,
	})

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
				node.RegisterStoreVaultServiceHandler,
			},
			CorsHandler: c.Handler,
			Swagger: &gateway.Swagger{
				HostURL:            s.Config.Swagger.HostURL,
				BasePath:           s.Config.Swagger.BasePath,
				SwaggerPath:        configs.SwaggerStoreVaultPath,
				FsSwagger:          swagger.FsSwaggerStoreVault,
				OpenAPIPath:        configs.SwaggerOpenAPIStoreVaultPath,
				FsOpenAPI:          third_party.OpenAPIStoreVault,
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
