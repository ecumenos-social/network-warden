package configurations

import (
	"github.com/ecumenos-social/network-warden/pkg/fxpostgres"
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	holdersessions "github.com/ecumenos-social/network-warden/services/holder-sessions"
	smssender "github.com/ecumenos-social/network-warden/services/sms-sender"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxidgenerator"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

type fxConfig struct {
	fx.Out

	App           *toolkitfx.AppConfig
	Logger        *fxlogger.Config
	GRPC          *fxgrpc.Config
	Postgres      *fxpostgres.Config
	IDGenerator   *fxidgenerator.Config
	Auth          *auth.Config
	HolderSession *holdersessions.Config
	Emailer       *emailer.Config
	SMSSender     *smssender.Config
}

var Module = func(cctx *cli.Context) fx.Option {
	return fx.Options(
		fx.Provide(func() fxConfig {
			return fxConfig{
				App: &toolkitfx.AppConfig{
					Name:        cctx.String("nw-app-name"),
					Description: cctx.String("nw-app-description"),
					RateLimit:   cctx.Float64("nw-app-rate-limit"),
				},
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
				IDGenerator: &fxidgenerator.Config{
					TopNodeID: cctx.Int64("nw-id-gen-top-node-id"),
					LowNodeID: cctx.Int64("nw-id-gen-low-node-id"),
				},
				Auth: &auth.Config{
					JWTSigningKey:   cctx.String("nw-auth-jwt-signing-key"),
					TokenAge:        cctx.Duration("nw-auth-token-age"),
					RefreshTokenAge: cctx.Duration("nw-auth-refresh-token-age"),
				},
				HolderSession: &holdersessions.Config{
					Age: cctx.Duration("nw-holder-session-age"),
				},
				Emailer:   &emailer.Config{},
				SMSSender: &smssender.Config{},
			}
		}),
	)
}
