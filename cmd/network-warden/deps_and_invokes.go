package main

import (
	"github.com/ecumenos-social/network-warden/cmd/network-warden/configurations"
	"github.com/ecumenos-social/network-warden/cmd/network-warden/grpc"
	"github.com/ecumenos-social/network-warden/pkg/fxpostgres"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	"go.uber.org/fx"
)

var Dependencies = fx.Options(
	fx.Supply(toolkitfx.ServiceName(configurations.ServiceName)),
	fxlogger.Module,
	fx.Provide(
		grpc.NewGRPCServer,
		grpc.NewHTTPGatewayHandler,
		grpc.NewLivenessGateway,
	),
	fxpostgres.Module,
)

var Invokes = fx.Invoke(
	fxgrpc.RunRegisteredGRPCServer,
	grpc.RunHTTPGateway,
	fxgrpc.RunHealthServer,
	fxgrpc.RunLivenessGateway,
)
