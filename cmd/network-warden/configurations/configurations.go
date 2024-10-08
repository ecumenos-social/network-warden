package configurations

import (
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/network-warden/services/jwt"
	smssender "github.com/ecumenos-social/network-warden/services/sms-sender"
	"github.com/ecumenos-social/toolkit/types"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	"github.com/ecumenos-social/toolkitfx/fxpostgres"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

type fxConfig struct {
	fx.Out

	App                          *toolkitfx.GenericAppConfig
	AppSpecific                  *toolkitfx.NetworkWardenAppConfig
	Logger                       *fxlogger.Config
	GRPC                         *fxgrpc.Config
	Postgres                     *fxpostgres.Config
	HolderSessionsIDGenerator    *idgenerators.HolderSessionsIDGeneratorConfig
	HoldersIDGenerator           *idgenerators.HoldersIDGeneratorConfig
	SentEmailsIDGenerator        *idgenerators.SentEmailsIDGeneratorConfig
	NetworkNodesIDGenerator      *idgenerators.NetworkNodesIDGeneratorConfig
	PersonalDataNodesIDGenerator *idgenerators.PersonalDataNodesIDGeneratorConfig
	NetworkWardensIDGenerator    *idgenerators.NetworkWardensIDGeneratorConfig
	JWT                          *jwt.Config
	Auth                         *auth.Config
	Emailer                      *emailer.Config
	SMSSender                    *smssender.Config
}

var Module = func(cctx *cli.Context) fx.Option {
	return fx.Options(
		fx.Provide(func() fxConfig {
			return fxConfig{
				App: &toolkitfx.GenericAppConfig{
					ID:          cctx.Int64("nw-app-id"),
					IDGenNode:   cctx.Int64("nw-app-id-gen-node"),
					Name:        cctx.String("nw-app-name"),
					Description: cctx.String("nw-app-description"),
					RateLimit: &types.RateLimit{
						MaxRequests: cctx.Int64("nw-app-rate-limit-max-requests"),
						Interval:    cctx.Duration("nw-app-rate-limit-interval"),
					},
				},
				AppSpecific: &toolkitfx.NetworkWardenAppConfig{
					AddressSuffix: cctx.String("nw-app-address-suffix"),
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
				HolderSessionsIDGenerator: &idgenerators.HolderSessionsIDGeneratorConfig{
					TopNodeID: cctx.Int64("nw-app-id-gen-node"),
					LowNodeID: 0,
				},
				HoldersIDGenerator: &idgenerators.HoldersIDGeneratorConfig{
					TopNodeID: cctx.Int64("nw-app-id-gen-node"),
					LowNodeID: 0,
				},
				SentEmailsIDGenerator: &idgenerators.SentEmailsIDGeneratorConfig{
					TopNodeID: cctx.Int64("nw-app-id-gen-node"),
					LowNodeID: 0,
				},
				NetworkNodesIDGenerator: &idgenerators.NetworkNodesIDGeneratorConfig{
					TopNodeID: cctx.Int64("nw-app-id-gen-node"),
					LowNodeID: 0,
				},
				PersonalDataNodesIDGenerator: &idgenerators.PersonalDataNodesIDGeneratorConfig{
					TopNodeID: cctx.Int64("nw-app-id-gen-node"),
					LowNodeID: 0,
				},
				NetworkWardensIDGenerator: &idgenerators.NetworkWardensIDGeneratorConfig{
					TopNodeID: cctx.Int64("nw-app-id-gen-node"),
					LowNodeID: 0,
				},
				JWT: &jwt.Config{
					SigningKey:      cctx.String("nw-jwt-signing-key"),
					TokenAge:        cctx.Duration("nw-jwt-token-age"),
					RefreshTokenAge: cctx.Duration("nw-jwt-refresh-token-age"),
				},
				Auth: &auth.Config{
					SessionAge: cctx.Duration("nw-auth-session-age"),
				},
				Emailer: &emailer.Config{
					SMTPHost:           cctx.String("nw-emailer-smtp-host"),
					SMTPPort:           cctx.String("nw-emailer-smtp-port"),
					SenderUsername:     cctx.String("nw-emailer-sender-username"),
					SenderEmailAddress: cctx.String("nw-emailer-sender-email-address"),
					SenderPassword:     cctx.String("nw-emailer-sender-password"),
					ConfirmationOfRegistration: &emailer.RateLimit{
						MaxRequests: cctx.Int64("nw-emailer-confirmation-of-registration-max-requests"),
						Interval:    cctx.Duration("nw-emailer-confirmation-of-registration-interval"),
					},
				},
				SMSSender: &smssender.Config{},
			}
		}),
	)
}
