package main

import (
	"github.com/ecumenos-social/network-warden/cmd/network-warden/configurations"
	"github.com/ecumenos-social/network-warden/cmd/network-warden/grpc"
	"github.com/ecumenos-social/network-warden/pkg/toolkitfx"
	"github.com/ecumenos-social/network-warden/pkg/toolkitfx/fxenvironment"
	"github.com/ecumenos-social/network-warden/pkg/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/network-warden/pkg/toolkitfx/fxlogger"
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
		grpc.NewGatewayHandler,
		grpc.NewLivenessGateway,
	),
)

var Invokes = fx.Invoke(
	fxgrpc.NewRegisteredGRPCServer,
	grpc.NewHTTPGateway,
	fxgrpc.NewHealthServer,
	fxgrpc.NewLivenessGateway,
)
