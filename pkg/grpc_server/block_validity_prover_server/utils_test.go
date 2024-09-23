package block_validity_prover_server_test

import (
	"context"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/docs/swagger"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gateway"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	node "intmax2-node/internal/pb/gen/block_validity_prover_service/node"
	"intmax2-node/internal/pb/listener"
	server "intmax2-node/pkg/grpc_server/block_validity_prover_server"
	"intmax2-node/third_party"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/dimiro1/health"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

func Start(
	commands server.Commands,
	ctx context.Context,
	cfg *configs.Config,
	log logger.Logger,
	dbApp server.SQLDriverApp,
	hc *health.Handler,
	sb server.ServiceBlockchain,
) (gRPCServerStop func(), gwServer *http.Server) {
	s := httptest.NewServer(nil)
	s.Close()

	s2 := httptest.NewServer(nil)
	s2.Close()

	const httpSplitter = "http://"

	addr := strings.TrimPrefix(s.URL, httpSplitter)

	tm := time.Duration(cfg.HTTP.Timeout) * time.Second

	c := cors.New(cors.Options{
		AllowedOrigins:       cfg.HTTP.CORS,
		AllowedMethods:       cfg.HTTP.CORSAllowMethods,
		AllowedHeaders:       cfg.HTTP.CORSAllowHeaders,
		ExposedHeaders:       cfg.HTTP.CORSExposeHeaders,
		AllowCredentials:     cfg.HTTP.CORSAllowCredentials,
		MaxAge:               cfg.HTTP.CORSMaxAge,
		OptionsSuccessStatus: cfg.HTTP.CORSStatusCode,
	})

	srv := server.New(log, cfg, dbApp, commands, cfg.HTTP.CookieForAuthUse, hc, sb)
	ctx = context.WithValue(ctx, consts.AppConfigs, cfg)

	const (
		app       = "app"
		appName   = " (node) "
		version   = "version"
		buildtime = "buildtime"
	)

	// run externals gRPC server listener
	_, gRPCServerStop = listener.Run(
		ctx,
		log,
		appName,
		addr, // listen incoming host:port for gRPC server
		func(sr grpc.ServiceRegistrar) {
			node.RegisterInfoServiceServer(sr, srv)
		},
	)

	// healthCheck
	hc.AddInfo(app, map[string]any{
		version:   buildvars.Version,
		buildtime: buildvars.BuildTime,
	})

	// run web -> gRPC gateway
	gw, _ := gateway.Run(
		ctx,
		&gateway.Params{
			Name:               appName,
			Logger:             log,
			GatewayAddr:        strings.TrimPrefix(s2.URL, httpSplitter), // listen incoming host:port for rest api
			DialAddr:           addr,                                     // connect to gRPC server host:port
			HTTPTimeout:        tm,
			HealthCheckHandler: hc,
			Services: []gateway.RegisterServiceHandlerFunc{
				node.RegisterInfoServiceHandler,
			},
			CorsHandler: c.Handler,
			Swagger: &gateway.Swagger{
				HostURL:            cfg.Swagger.HostURL,
				BasePath:           cfg.Swagger.BasePath,
				SwaggerPath:        configs.SwaggerBlockValidityProverPath,
				FsSwagger:          swagger.FsSwaggerBlockValidityProver,
				OpenAPIPath:        configs.SwaggerOpenAPIBlockValidityProverPath,
				FsOpenAPI:          third_party.OpenAPIBlockValidityProver,
				RegexpBuildVersion: cfg.Swagger.RegexpBuildVersion,
				RegexpHostURL:      cfg.Swagger.RegexpHostURL,
				RegexpBasePATH:     cfg.Swagger.RegexpBasePATH,
			},
			Cookies: &http_response_modifier.Cookies{
				ForAuthUse:         cfg.HTTP.CookieForAuthUse,
				Secure:             cfg.HTTP.CookieSecure,
				Domain:             cfg.HTTP.CookieDomain,
				SameSiteStrictMode: cfg.HTTP.CookieSameSiteStrictMode,
			},
		},
	)

	gwServer = gw.Server()

	return gRPCServerStop, gwServer
}
