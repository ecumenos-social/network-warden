package main

import (
	"github.com/ecumenos-social/network-warden/cmd/admin/configurations"
	"github.com/ecumenos-social/network-warden/cmd/admin/grpc"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxenvironment"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	"go.uber.org/fx"
)

type fxConfig struct {
	fx.Out
	Logger    fxlogger.Config
	Grpc      fxgrpc.Config
	GrpcLocal grpc.Config
}

var Dependencies = fx.Options(
	fx.Supply(toolkitfx.ServiceName(configurations.ServiceName)),
	fxlogger.Module,
	fxenvironment.Module(&fxConfig{}, false),
	fx.Provide(func(c fxenvironment.FxConfig) fxConfig {
		return *c.(*fxConfig)
	}),
	fx.Provide(
		grpc.NewGRPCServer,
		// TODO: uncomment when endpoints are added
		// grpc.NewGatewayHandler,
		// grpc.NewLivenessGateway,
	),
)

var Invokes = fx.Invoke(
	fxgrpc.NewRegisteredGRPCServer,
	// TODO: uncomment when endpoints are added
	// grpc.NewHTTPGateway,
	// fxgrpc.NewHealthServer,
	// fxgrpc.NewLivenessGateway,
)
