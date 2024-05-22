package configurations

import (
	"github.com/ecumenos-social/network-warden/pkg/fxpostgres"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

type fxConfig struct {
	fx.Out

	Logger   *fxlogger.Config
	GRPC     *fxgrpc.Config
	Postgres *fxpostgres.Config
}

var Module = func(cctx *cli.Context) fx.Option {
	return fx.Options(
		fx.Provide(func() fxConfig {
			return fxConfig{
				Logger: &fxlogger.Config{
					Production: cctx.Bool("nw-logger-production"),
				},
				GRPC: &fxgrpc.Config{
					GRPC: fxgrpc.GRPCConfig{
						Host:                                    cctx.String("nw-grpc-host"),
						Port:                                    cctx.String("nw-grpc-port"),
						MaxConnectionAge:                        cctx.Duration("nw-grpc-max-conn-age"),
						KeepAliveEnforcementMinTime:             cctx.Duration("nw-grpc-keep-alive-enforcement-min-time"),
						KeepAliveEnforcementPermitWithoutStream: cctx.Bool("nw-grpc-keep-alive-enforcement-permit-without-stream"),
					},
					Health: fxgrpc.HealthConfig{
						Enabled: cctx.Bool("nw-enabled-health-server"),
						Host:    cctx.String("nw-health-server-host"),
						Port:    cctx.String("nw-health-server-port"),
					},
					HTTPGateway: fxgrpc.HTTPGatewayConfig{
						Host: cctx.String("nw-http-gateway-host"),
						Port: cctx.String("nw-http-gateway-port"),
					},
					LivenessGateway: fxgrpc.LivenessGatewayConfig{
						Host: cctx.String("nw-liveness-gateway-host"),
						Port: cctx.String("nw-liveness-gateway-port"),
					},
				},
				Postgres: &fxpostgres.Config{
					URL:            cctx.String("nw-postgres-url"),
					MigrationsPath: cctx.String("nw-postgres-migrations-path"),
				},
			}
		}),
	)
}
