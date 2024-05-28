package grpc

import (
	"context"
	"net"
	"net/http"
	"time"

	grpcutils "github.com/ecumenos-social/grpc-utils"
	pbv1 "github.com/ecumenos-social/schemas/proto/gen/networkwarden/v1"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/heptiolabs/healthcheck"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type grpcServerParams struct {
	fx.In

	LC          fx.Lifecycle
	Config      *fxgrpc.Config
	ServiceName toolkitfx.ServiceName
	Handler     *Handler
}

func NewGRPCServer(params grpcServerParams) *fxgrpc.GRPCServer {
	grpcServer := fxgrpc.GRPCServer{}
	params.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			server := grpcutils.NewServer(string(params.ServiceName), net.JoinHostPort(params.Config.GRPC.Host, params.Config.GRPC.Port))
			server.Init(
				grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
					MinTime:             params.Config.GRPC.KeepAliveEnforcementMinTime,
					PermitWithoutStream: params.Config.GRPC.KeepAliveEnforcementPermitWithoutStream,
				}),
				grpcutils.ValidatorServerOption(),
				grpcutils.RecoveryServerOption(),
				grpc.KeepaliveParams(keepalive.ServerParameters{MaxConnectionAge: params.Config.GRPC.MaxConnectionAge}),
			)
			pbv1.RegisterNetworkWardenServiceServer(server.Server, params.Handler)
			grpcServer.Server = server

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return &grpcServer
}

func NewHTTPGatewayHandler() *fxgrpc.HTTPGatewayHandler {
	return &fxgrpc.HTTPGatewayHandler{
		Handler: pbv1.RegisterNetworkWardenServiceHandler,
	}
}

func NewLivenessGateway() *fxgrpc.LivenessGatewayHandler {
	health := healthcheck.NewHandler()
	health.AddLivenessCheck("healthcheck", func() error { return nil })
	return &fxgrpc.LivenessGatewayHandler{Handler: health}
}

func RunHTTPGateway(
	lc fx.Lifecycle,
	s fx.Shutdowner,
	logger *zap.Logger,
	cfg *fxgrpc.Config,
	g *fxgrpc.HTTPGatewayHandler,
) error {
	httpAddr := net.JoinHostPort(cfg.HTTPGateway.Host, cfg.HTTPGateway.Port)
	mux := runtime.NewServeMux()
	conn := grpcutils.NewClientConnection(net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port))

	zapConf := zap.NewProductionConfig()
	zapConf.Level.SetLevel(zap.ErrorLevel)
	errLogger, err := zapConf.Build()
	if err != nil {
		errLogger = logger.With()
	}

	_ = conn.Dial(grpcutils.DefaultDialOpts(errLogger)...)
	if err := g.Handler(context.Background(), mux, conn.Connection); err != nil {
		logger.Error("failed to register mapping service handler", zap.Error(err))
	}

	var httpServer *http.Server
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				httpServer = &http.Server{Addr: httpAddr, Handler: mux}
				logger.Info("starting HTTP gateway...", zap.String("addr", httpAddr))
				err = httpServer.ListenAndServe()
				if err != nil {
					logger.Error("failed to start http server", zap.Error(err))
					_ = s.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			_ = conn.CleanUp()
			if httpServer != nil {
				timeout, can := context.WithTimeout(context.Background(), 10*time.Second)
				defer can()
				if err := httpServer.Shutdown(timeout); err != nil {
					logger.Error("stopped http server after gRPC failure", zap.Error(err))
				}
			}
			return nil
		},
	})

	return nil
}
