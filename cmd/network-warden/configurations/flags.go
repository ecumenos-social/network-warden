package configurations

import (
	"time"

	cli "github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
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
	&cli.Float64Flag{
		Name:    "nw-app-rate-limit",
		Usage:   "it is rate limit",
		Value:   0.1,
		EnvVars: []string{"NETWORK_WARDEN_APP_RATE_LIMIT"},
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
		Value:   `file://cmd/network-warden/pgmigrations`,
		EnvVars: []string{"NETWORK_WARDEN_POSTGRES_MIGRATIONS_PATH"},
	},
	&cli.Int64Flag{
		Name:    "nw-id-gen-top-node-id",
		Usage:   "it is id generator top level node seed",
		Value:   10,
		EnvVars: []string{"NETWORK_WARDEN_ID_GENERATOR_TOP_NODE_ID"},
	},
	&cli.Int64Flag{
		Name:    "nw-id-gen-low-node-id",
		Usage:   "it is id generator low level node seed",
		Value:   10,
		EnvVars: []string{"NETWORK_WARDEN_ID_GENERATOR_LOW_NODE_ID"},
	},
	&cli.StringFlag{
		Name:    "nw-auth-jwt-signing-key",
		Usage:   "it is JWT secret",
		Value:   "alDFsk1d2!j@G$4%5^B&f*6(7)h_-g+=",
		EnvVars: []string{"NETWORK_WARDEN_AUTH_JWT_SIGNING_KEY"},
	},
	&cli.DurationFlag{
		Name:    "nw-auth-token-age",
		Usage:   "it is age of token",
		Value:   30 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_AUTH_TOKEN_AGE"},
	},
	&cli.DurationFlag{
		Name:    "nw-auth-refresh-token-age",
		Usage:   "it is age of refresh token",
		Value:   90 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_AUTH_REFRESH_TOKEN_AGE"},
	},
	&cli.DurationFlag{
		Name:    "nw-holder-session-age",
		Usage:   "it is age of holder session",
		Value:   90 * time.Minute,
		EnvVars: []string{"NETWORK_WARDEN_HOLDER_SESSION_AGE"},
	},
}
