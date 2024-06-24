package gateway

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/configs/buildvars"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gateway/consts"
	"intmax2-node/internal/pb/gateway/http_request_modifier"
	"intmax2-node/internal/pb/gateway/http_response_modifier"
	"io/fs"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/dimiro1/health"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
)

// routes
const (
	openAPIPath    = "/swagger/"
	prometheusPath = "/prometheus"
	healthPath     = "/health"
	statusPath     = "/status"
)

// RegisterServiceHandlerFunc func to register gRPC service handler.
type RegisterServiceHandlerFunc func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

type Swagger struct {
	HostURL  string
	BasePath string

	SwaggerPath string
	FsSwagger   embed.FS

	OpenAPIPath string
	FsOpenAPI   embed.FS

	RegexpBuildVersion *regexp.Regexp
	RegexpHostURL      *regexp.Regexp
	RegexpBasePATH     *regexp.Regexp
}

// Params describes parameters for grpc_gateway
type Params struct {
	Name                  string
	Logger                logger.Logger
	GatewayAddr, DialAddr string
	HTTPTimeout           time.Duration
	HealthCheckHandler    *health.Handler
	Services              []RegisterServiceHandlerFunc
	CorsHandler           func(http.Handler) http.Handler
	Swagger               *Swagger
	Cookies               *http_response_modifier.Cookies
}

type Gateway interface {
	SetStatus(status health.Status)
	GetStatus() health.Status
	Server() *http.Server
}

type gateway struct {
	config *Params
	status health.Status
	srv    *http.Server
}

// Run runs the gRPC-Gateway on the gatewayAddr using gRPC client connection
// as underlying gRPC client connection to gRPC server started before.
func Run(ctx context.Context, config *Params) (Gateway, chan error) { // nolint:gocritic
	l := config.Logger
	cfg := ctx.Value(consts.AppConfigs).(*configs.Config)

	errCh := make(chan error, 1)

	gw := gateway{
		config: config,
		status: health.Unknown,
	}

	interceptorOpt := otelgrpc.WithTracerProvider(otel.GetTracerProvider())
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(interceptorOpt)),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor(interceptorOpt)),
	}

	creds, err := tls.LoadX509KeyPair(cfg.APP.PEMPathClientCert, cfg.APP.PEMPathClientKey)
	if err != nil {
		log.Fatalf("load X509 key pair error: %+v", err)
	}

	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(cfg.APP.PEMPathCACert)
	if err != nil {
		log.Fatalf("load CA Cert error occurred: %+v", err)
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		log.Fatalf("append CA Cert from PEM to X509 cert pool error occurred")
	}

	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		ServerName:   cfg.APP.CADomainName,
		Certificates: []tls.Certificate{creds},
		RootCAs:      ca,
		MinVersion:   tls.VersionTLS13,
	})))

	grpcClient, err := grpc.DialContext(
		ctx,
		"dns:///"+gw.config.DialAddr,
		opts...,
	)
	if err != nil {
		const msg = "%sfailed to connect to gRPC-server: %w"
		errCh <- fmt.Errorf(msg, gw.config.Name, err)
		return nil, errCh
	}

	httpRespM := http_response_modifier.NewProcessing(gw.config.Cookies)
	httpErrM := http_response_modifier.NewHTTPErrorHandler(httpRespM, gw.config.Cookies)

	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(http_request_modifier.Middleware),
		runtime.WithForwardResponseOption(httpRespM.Middleware),
		runtime.WithErrorHandler(httpErrM.HTTPErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	// register gRPC services on gwMux
	for _, fn := range config.Services {
		err = fn(ctx, gwMux, grpcClient)
		if err != nil {
			const msg = "%sfailed to register gateway: %w"
			errCh <- fmt.Errorf(msg, gw.config.Name, err)
			return nil, errCh
		}
	}

	// prepare OpenAPI handler
	oa := gw.getOpenAPIHandler()

	loggerMw := logRequest(config.Logger)

	swaggerURI := "/" + gw.config.Swagger.SwaggerPath
	jsonSwagger, err := gw.swaggerJSON()
	if err != nil {
		const msg = "%sfailed to get to swagger JSON: %w"
		errCh <- fmt.Errorf(msg, gw.config.Name, err)
		return nil, errCh
	}

	const readHeaderTimeout = 300 * time.Millisecond
	gwServer := &http.Server{
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: readHeaderTimeout,
		Addr:              config.GatewayAddr,
		Handler: config.CorsHandler(loggerMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// replace context by context with timeout
			rctx, cancel := context.WithTimeout(
				context.WithValue(r.Context(), consts.AppConfigs, cfg),
				config.HTTPTimeout,
			)
			defer cancel()
			r = r.WithContext(rctx)

			// GET /swagger/
			if strings.HasPrefix(r.URL.Path, openAPIPath) {
				http.StripPrefix(openAPIPath, oa).ServeHTTP(w, r)
				return
			}

			err = gwMux.HandlePath(http.MethodGet, healthPath, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				config.HealthCheckHandler.ServeHTTP(w, r)
			})

			err = gwMux.HandlePath(http.MethodGet, statusPath, gw.statusHandler())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = gwMux.HandlePath(http.MethodGet, swaggerURI, gw.swaggerHandler(jsonSwagger))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			err = gwMux.HandlePath(http.MethodGet, prometheusPath, func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
				promhttp.Handler().ServeHTTP(w, r)
			})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			gwMux.ServeHTTP(w, r)
		}))),
	}

	const (
		httpKey           = "http://"
		httpNameKey       = "HTTP"
		httpsKey          = "https://"
		httpsNameKey      = "HTTPS"
		servingGW         = "%sServing %s on %s%s"
		servingStatus     = "%sServing status on %s%s%s"
		servingHealth     = "%sServing health on %s%s%s"
		servingPrometheus = "%sServing prometheus metric on %s%s%s"
		servingOAPI       = "%sServing OpenAPI Documentation on %s%s%s"
		servingJSON       = "%sServing JSON OpenAPI Documentation on %s%s%s"
	)
	schema := httpKey
	serveType := httpNameKey
	if cfg.HTTP.TLSUse {
		schema = httpsKey
		serveType = httpsNameKey
	}
	l.Infof(servingGW, gw.config.Name, serveType, schema, gw.config.GatewayAddr)
	l.Infof(servingStatus, gw.config.Name, schema, config.GatewayAddr, statusPath)
	l.Infof(servingHealth, gw.config.Name, schema, config.GatewayAddr, healthPath)
	l.Infof(servingPrometheus, gw.config.Name, schema, config.GatewayAddr, prometheusPath)
	l.Infof(servingOAPI, gw.config.Name, schema, config.GatewayAddr, openAPIPath)
	l.Infof(servingJSON, gw.config.Name, schema, config.GatewayAddr, swaggerURI)
	go func() {
		if cfg.HTTP.TLSUse {
			gwServer.TLSConfig = &tls.Config{
				ServerName:   cfg.APP.CADomainName,
				Certificates: []tls.Certificate{creds},
				RootCAs:      ca,
				MinVersion:   tls.VersionTLS13,
			}
			errCh <- gwServer.ListenAndServeTLS(cfg.APP.PEMPathClientCert, cfg.APP.PEMPathClientKey)
		} else {
			errCh <- gwServer.ListenAndServe()
		}
		if err = gwServer.Shutdown(ctx); err != nil {
			const msg = "%sShutdown gRPC-Gateway error: %s"
			l.Errorf(msg, gw.config.Name, err)
		} else {
			const msg = "%sShutdown gRPC-Gateway"
			l.Infof(msg, gw.config.Name)
		}
	}()

	gw.srv = gwServer
	gw.status = health.Up

	return &gw, errCh
}

// getOpenAPIHandler serves an OpenAPI UI.
func (gw *gateway) getOpenAPIHandler() http.Handler {
	const (
		ext  = ".svg"
		typ  = "image/svg+xml"
		code = -1
	)
	_ = mime.AddExtensionType(ext, typ)
	// Use subdirectory in embedded files
	subFS, err := fs.Sub(gw.config.Swagger.FsOpenAPI, gw.config.Swagger.OpenAPIPath)
	if err != nil {
		const msg = "couldn't create sub filesystem: %s"
		gw.config.Logger.Errorf(msg, err)
		os.Exit(code)
	}
	return http.FileServer(http.FS(subFS))
}

func (gw *gateway) swaggerJSON() ([]byte, error) {
	return gw.config.Swagger.FsSwagger.ReadFile(gw.config.Swagger.SwaggerPath)
}

func (gw *gateway) swaggerHandler(payload []byte) runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(
			gw.config.Swagger.RegexpBasePATH.ReplaceAll(
				gw.config.Swagger.RegexpHostURL.ReplaceAll(
					gw.config.Swagger.RegexpBuildVersion.ReplaceAll(
						payload,
						[]byte(buildvars.Version),
					),
					[]byte(gw.config.Swagger.HostURL),
				),
				[]byte(gw.config.Swagger.BasePath),
			))
	}
}

func (gw *gateway) statusHandler() runtime.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(gw.status))
	}
}

func (gw *gateway) SetStatus(status health.Status) {
	gw.status = status
}

func (gw *gateway) GetStatus() health.Status {
	return gw.status
}

func (gw *gateway) Server() *http.Server {
	return gw.srv
}
