package configurations

import (
	"time"

	cli "github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.Int64Flag{
		Name:    "nw-app-id",
		Usage:   "it is application ID",
		Value:   0,
		EnvVars: []string{"NETWORK_WARDEN_APP_ID"},
	},
	&cli.Int64Flag{
		Name:    "nw-app-id-gen-node",
		Usage:   "it is application ID generation node",
		Value:   0,
		EnvVars: []string{"NETWORK_WARDEN_APP_ID_GEN_NODE"},
	},
	&cli.StringFlag{
		Name:    "nw-app-name",
		Usage:   "it is unique application name",
		Value:   "name",
		EnvVars: []string{"NETWORK_WARDEN_APP_NAME"},
	},
	&cli.StringFlag{
		Name:    "nw-app-description",
		Usage:   "it is application description",
		Value:   "it is network warden",
		EnvVars: []string{"NETWORK_WARDEN_APP_DESCRIPTION"},
	},
	&cli.StringFlag{
		Name:    "nw-app-address-suffix",
		Usage:   "it is unique address suffix specific for this network warden",
		Value:   "nw",
		EnvVars: []string{"NETWORK_WARDEN_APP_ADDRESS_SUFFIX"},
	},
	&cli.Int64Flag{
		Name:    "nw-app-rate-limit-max-requests",
		Usage:   "it is rate limit max requests amount",
		Value:   1000,
		EnvVars: []string{"NETWORK_WARDEN_APP_RATE_LIMIT_MAX_REQUESTS"},
	},
	&cli.DurationFlag{
		Name:    "nw-app-rate-limit-interval",
		Usage:   "it is rate limit interval",
		Value:   time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_APP_RATE_LIMIT_INTERVAL"},
	},
	&cli.BoolFlag{
		Name:    "nw-logger-production",
		Usage:   "make it true if you need logging on production environment",
		Value:   false,
		EnvVars: []string{"NETWORK_WARDEN_LOGGER_PRODUCTION"},
	},
	&cli.StringFlag{
		Name:    "nw-grpc-host",
		Usage:   "it is gRPC server host",
		Value:   "0.0.0.0",
		EnvVars: []string{"NETWORK_WARDEN_GRPC_HOST"},
	},
	&cli.StringFlag{
		Name:    "nw-grpc-port",
		Usage:   "it is gRPC server port",
		Value:   "8080",
		EnvVars: []string{"NETWORK_WARDEN_GRPC_PORT"},
	},
	&cli.StringFlag{
		Name:    "nw-http-gateway-host",
		Usage:   "it is HTTP gateway host",
		Value:   "0.0.0.0",
		EnvVars: []string{"NETWORK_WARDEN_HTTP_GATEWAY_HOST"},
	},
	&cli.StringFlag{
		Name:    "nw-http-gateway-port",
		Usage:   "it is HTTP gateway port",
		Value:   "9090",
		EnvVars: []string{"NETWORK_WARDEN_HTTP_GATEWAY_PORT"},
	},
	&cli.BoolFlag{
		Name:    "nw-enabled-health-server",
		Usage:   "make it true if you need to enable health server",
		Value:   false,
		EnvVars: []string{"NETWORK_WARDEN_ENABLED_HEALTH_SERVER"},
	},
	&cli.StringFlag{
		Name:    "nw-health-server-host",
		Usage:   "it is health server host",
		Value:   "0.0.0.0",
		EnvVars: []string{"NETWORK_WARDEN_HEALTH_SERVER_HOST"},
	},
	&cli.StringFlag{
		Name:    "nw-health-server-port",
		Usage:   "it is health server port",
		Value:   "10010",
		EnvVars: []string{"NETWORK_WARDEN_HEALTH_SERVER_PORT"},
	},
	&cli.StringFlag{
		Name:    "nw-liveness-gateway-host",
		Usage:   "it is liveness gateway host",
		Value:   "0.0.0.0",
		EnvVars: []string{"NETWORK_WARDEN_LIVENESS_GATEWAY_HOST"},
	},
	&cli.StringFlag{
		Name:    "nw-liveness-gateway-port",
		Usage:   "it is liveness gateway port",
		Value:   "8086",
		EnvVars: []string{"NETWORK_WARDEN_LIVENESS_GATEWAY_PORT"},
	},
	&cli.DurationFlag{
		Name:    "nw-grpc-max-conn-age",
		Usage:   "it is max age of connection with gRPC server",
		Value:   5 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_GRPC_MAX_CONNECTION_AGE"},
	},
	&cli.DurationFlag{
		Name:    "nw-grpc-keep-alive-enforcement-min-time",
		Usage:   "it is minimal time of keep alive enforcement gRPC server",
		Value:   time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_GRPC_KEEP_ALIVE_ENFORCEMENT_MIN_TIME"},
	},
	&cli.BoolFlag{
		Name:    "nw-grpc-keep-alive-enforcement-permit-without-stream",
		Usage:   "",
		Value:   true,
		EnvVars: []string{"NETWORK_WARDEN_GRPC_KEEP_ALIVE_ENFORCEMENT_PERMIT_WITHOUT_STREAM"},
	},
	&cli.StringFlag{
		Name:    "nw-postgres-url",
		Usage:   "it is URL of postgres database connected to the app",
		Value:   `postgresql://ecumenosuser:rootpassword@localhost:5432/ecumenos_network_warden_db`,
		EnvVars: []string{"NETWORK_WARDEN_POSTGRES_URL"},
	},
	&cli.StringFlag{
		Name:    "nw-postgres-migrations-path",
		Usage:   "it is path to directory with postgres migrations",
		Value:   `file://pgmigrations`,
		EnvVars: []string{"NETWORK_WARDEN_POSTGRES_MIGRATIONS_PATH"},
	},
	&cli.StringFlag{
		Name:    "nw-jwt-signing-key",
		Usage:   "it is JWT secret",
		Value:   "alDFsk1d2!j@G$4%5^B&f*6(7)h_-g+=",
		EnvVars: []string{"NETWORK_WARDEN_JWT_SIGNING_KEY"},
	},
	&cli.DurationFlag{
		Name:    "nw-jwt-token-age",
		Usage:   "it is age of token",
		Value:   30 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_JWT_TOKEN_AGE"},
	},
	&cli.DurationFlag{
		Name:    "nw-jwt-refresh-token-age",
		Usage:   "it is age of refresh token",
		Value:   90 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_JWT_REFRESH_TOKEN_AGE"},
	},
	&cli.DurationFlag{
		Name:    "nw-auth-session-age",
		Usage:   "it is age of holder session",
		Value:   90 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_AUTH_SESSION_AGE"},
	},
	&cli.StringFlag{
		Name:    "nw-emailer-smtp-host",
		Usage:   "it is SMTP server host",
		Value:   "smtp.sendgrid.net",
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_SMTP_HOST"},
	},
	&cli.StringFlag{
		Name:  "nw-emailer-smtp-port",
		Usage: "it is SMTP server port",
		Value: "465",
		// 25, 587	(for unencrypted/TLS connections)
		// 465	(for SSL connections)
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_SMTP_PORT"},
	},
	&cli.StringFlag{
		Name:    "nw-emailer-sender-username",
		Usage:   "it is emailer sender's username. It is needed for authentication",
		Value:   "apikey",
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_SENDER_USERNAME"},
	},
	&cli.StringFlag{
		Name:    "nw-emailer-sender-password",
		Usage:   "it is emailer sender's password. It is needed for authentication",
		Value:   "secret-apikey",
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_SENDER_PASSWORD"},
	},
	&cli.StringFlag{
		Name:    "nw-emailer-sender-email-address",
		Usage:   "it is emailer sender's email address. It is needed for putting it to message",
		Value:   "example@mail.com",
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_SENDER_EMAIL_ADDRESS"},
	},
	&cli.Int64Flag{
		Name:    "nw-emailer-confirmation-of-registration-max-requests",
		Usage:   "it is rate limit value for maximal amount of requests for some interval",
		Value:   3,
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_CONFIRMATION_OF_REGISTRATION_MAX_REQUESTS"},
	},
	&cli.DurationFlag{
		Name:    "nw-emailer-confirmation-of-registration-interval",
		Usage:   "it is rate limit value for interval when we measure requests",
		Value:   5 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_EMAILER_CONFIRMATION_OF_REGISTRATION_INTERVAL"},
	},
}
