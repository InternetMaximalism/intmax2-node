package listener

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"intmax2-node/configs"
	"intmax2-node/internal/logger"
	"intmax2-node/internal/pb/gateway/consts"
	"net"
	"os"

	mw "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	throttle "github.com/yaronsumel/grpc-throttle"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

type RegisterServiceCallback func(s grpc.ServiceRegistrar)

func Run(
	ctx context.Context,
	log logger.Logger,
	name, addr string,
	callback RegisterServiceCallback,
) (errCh chan error, stop func()) {
	lc := net.ListenConfig{}
	cfg := ctx.Value(consts.AppConfigs).(*configs.Config)

	lis, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		log.Fatalf("%sbind port for gRPC server on %s", name, addr)
	}

	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair(cfg.APP.PEMPathServCert, cfg.APP.PEMPathServKey)
	if err != nil {
		log.Fatalf("%sload X509 key pair error: %+v", name, err)
	}

	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(cfg.APP.PEMPAthCACertClient)
	if err != nil {
		log.Fatalf("%sload CA Cert Client error occurred: %+v", name, err)
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		log.Fatalf("%sappend CA Cert Client from PEM to X509 cert pool error occurred", name)
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p any) (err error) {
			return status.Errorf(codes.Unknown, "%spanic triggered: %v", name, p)
		}),
	}

	prm := prometheus.NewServerMetrics()

	interceptorOpt := otelgrpc.WithTracerProvider(otel.GetTracerProvider())

	s := grpc.NewServer(
		grpc.StreamInterceptor(mw.ChainStreamServer(
			prm.StreamServerInterceptor(),
			otelgrpc.StreamServerInterceptor(interceptorOpt),
			// keep it last in the interceptor chain
			throttle.StreamServerInterceptor(throttleFn),
		)),
		grpc.UnaryInterceptor(mw.ChainUnaryServer(
			recovery.UnaryServerInterceptor(opts...),
			prm.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(interceptorOpt),
			// keep it last in the interceptor chain
			throttle.UnaryServerInterceptor(throttleFn),
		)),
		grpc.Creds(credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{cert},
			ClientCAs:    ca,
			MinVersion:   tls.VersionTLS13,
		})),
	)

	callback(s)

	prm.InitializeMetrics(s)

	errCh = make(chan error, 1)

	const servingGRPC = "%sServing gRPC server on %s"
	log.Infof(servingGRPC, name, addr)
	go func() {
		errCh <- s.Serve(lis)
		log.Infof("%sShutdown gRPC server.", name)
	}()

	return errCh, s.Stop
}
