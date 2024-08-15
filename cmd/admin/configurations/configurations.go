package configurations

import (
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	"github.com/ecumenos-social/toolkitfx/fxpostgres"
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
					Production: cctx.Bool("nw-admin-logger-production"),
				},
				GRPC: &fxgrpc.Config{
					GRPC: fxgrpc.GRPCConfig{
						Host:                                    cctx.String("nw-admin-grpc-host"),
						Port:                                    cctx.String("nw-admin-grpc-port"),
						MaxConnectionAge:                        cctx.Duration("nw-admin-grpc-max-conn-age"),
						KeepAliveEnforcementMinTime:             cctx.Duration("nw-admin-grpc-keep-alive-enforcement-min-time"),
						KeepAliveEnforcementPermitWithoutStream: cctx.Bool("nw-admin-grpc-keep-alive-enforcement-permit-without-stream"),
					},
					Health: fxgrpc.HealthConfig{
						Enabled: cctx.Bool("nw-admin-enabled-health-server"),
						Host:    cctx.String("nw-admin-health-server-host"),
						Port:    cctx.String("nw-admin-health-server-port"),
					},
					HTTPGateway: fxgrpc.HTTPGatewayConfig{
						Host: cctx.String("nw-admin-http-gateway-host"),
						Port: cctx.String("nw-admin-http-gateway-port"),
					},
					LivenessGateway: fxgrpc.LivenessGatewayConfig{
						Host: cctx.String("nw-admin-liveness-gateway-host"),
						Port: cctx.String("nw-admin-liveness-gateway-port"),
					},
				},
				Postgres: &fxpostgres.Config{
					URL:            cctx.String("nw-admin-postgres-url"),
					MigrationsPath: cctx.String("nw-postgres-migrations-path"),
				},
			}
		}),
	)
}
