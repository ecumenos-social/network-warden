package main

import (
	"github.com/ecumenos-social/network-warden/cmd/admin/configurations"
	"github.com/ecumenos-social/network-warden/cmd/admin/grpc"
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
		// TODO: uncomment when endpoints are added
		// grpc.NewGatewayHandler,
		// grpc.NewLivenessGateway,
	),
)

var Invokes = fx.Invoke(
	fxgrpc.RunRegisteredGRPCServer,
	// TODO: uncomment when endpoints are added
	// grpc.RunHTTPGateway,
	// fxgrpc.RunHealthServer,
	// fxgrpc.RunLivenessGateway,
)
