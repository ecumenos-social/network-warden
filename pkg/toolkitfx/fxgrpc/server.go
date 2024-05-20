package fxgrpc

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/ecumenos-social/network-warden/pkg/grpcutils"
	"github.com/ecumenos-social/network-warden/pkg/toolkitfx"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Config struct {
	GRPC     GRPCConfig
	HTTP     HTTPConfig
	Health   HealthConfig
	Liveness LivenessConfig
}

type GRPCConfig struct {
	Host string `default:"0.0.0.0" json:"host" yaml:"host"`
	Port string `default:"8080" json:"port" yaml:"port"`
}

type HTTPConfig struct {
	Host string `default:"0.0.0.0" json:"host" yaml:"host"`
	Port string `default:"9090" json:"port" yaml:"port"`
}

type HealthConfig struct {
	Enabled bool   `default:"true" json:"enabled" yaml:"enabled"`
	Host    string `default:"0.0.0.0" json:"host" yaml:"host"`
	Port    string `default:"10010" json:"port" yaml:"port"`
}

type LivenessConfig struct {
	Host string `default:"0.0.0.0" json:"host" yaml:"host"`
	Port string `default:"8086" json:"port" yaml:"port"`
}

type GRPCServer struct {
	Server *grpcutils.Server
}

type GatewayHandler struct {
	Handler func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

type LivenessHandler struct {
	Handler http.Handler
}

func NewRegisteredGRPCServer(lc fx.Lifecycle, s fx.Shutdowner, logger *zap.Logger, gs *GRPCServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("starting gRPC server...")
				if err := gs.Server.Serve(); err != nil {
					logger.Error("registered gRPC server stopping down due to error", zap.Error(err))
					_ = s.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if err := gs.Server.CleanUp(); err != nil {
				logger.Error("can not clean up server resources", zap.Error(err))
			}
			return nil
		},
	})
}

func NewHealthServer(lc fx.Lifecycle, s fx.Shutdowner, cfg Config, logger *zap.Logger, sn toolkitfx.ServiceName) {
	addr := net.JoinHostPort(cfg.Health.Host, cfg.Health.Port)
	var healthServer *grpcutils.HealthHandler
	if cfg.Health.Enabled {
		healthServer = grpcutils.NewHealthServer(string(sn), net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port), logger, addr)

		localCtx, cancel := context.WithCancel(context.Background())

		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					logger.Info("starting  gRPC health server...", zap.String("addr", addr))
					err := healthServer.GServer.Serve()
					if err != nil {
						logger.Error("HealthServer gRPC server stopping down due to error", zap.Error(err))
						_ = s.Shutdown()
					}
				}()

				go func() {
					err := healthServer.CheckConnection(localCtx)
					if err != nil {
						logger.Error("HealthServer CheckConnection stopping down due to error", zap.Error(err))
						_ = s.Shutdown()
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				cancel()
				if err := healthServer.GServer.CleanUp(); err != nil {
					logger.Warn("Could not clean up Health server resources", zap.Error(err))
				}
				return nil
			},
		})
	}
}

func NewHTTPGateway(lc fx.Lifecycle, s fx.Shutdowner, logger *zap.Logger, cfg Config, g *GatewayHandler) error {
	addr := net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port)
	mux := runtime.NewServeMux()
	conn := grpcutils.NewClientConnection(net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port))

	var httpServer *http.Server
	localCtx := context.Background()
	_ = conn.Dial(grpcutils.DefaultDialOpts(logger)...)
	err := g.Handler(localCtx, mux, conn.Connection)
	if err != nil {
		logger.Error("Failed to register mapping service handler", zap.Error(err))
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("starting HTTP gateway", zap.String("addr", addr))
				httpServer = &http.Server{Addr: addr, Handler: mux}
				err = httpServer.ListenAndServe()
				if err != nil {
					logger.Error("Failed to run HTTP gateway", zap.Error(err))
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
					logger.Error("Stopped http server after grpc failure", zap.Error(err))
				}
			}
			return nil
		},
	})

	return nil
}

func NewLivenessGateway(lc fx.Lifecycle, s fx.Shutdowner, logger *zap.Logger, cfg Config, h *LivenessHandler) error {
	addr := net.JoinHostPort(cfg.Liveness.Host, cfg.Liveness.Port)
	var httpServer *http.Server

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("starting liveness gateway...", zap.String("addr", addr))
				httpServer = &http.Server{Addr: addr, Handler: h.Handler}
				err := httpServer.ListenAndServe()
				if err != nil {
					logger.Error("Failed to run liveness gateway", zap.Error(err))
					_ = s.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if httpServer != nil {
				timeout, can := context.WithTimeout(context.Background(), 10*time.Second)
				defer can()
				if err := httpServer.Shutdown(timeout); err != nil {
					logger.Error("Stopped http server after grpc failure", zap.Error(err))
				}
			}
			return nil
		},
	})

	return nil
}
