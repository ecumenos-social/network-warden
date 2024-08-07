package main

import (
	"github.com/ecumenos-social/network-warden/cmd/admin/configurations"
	"github.com/ecumenos-social/network-warden/cmd/admin/grpc"
	"github.com/ecumenos-social/network-warden/cmd/admin/repository"
	"github.com/ecumenos-social/network-warden/pkg/fxpostgres"
	"github.com/ecumenos-social/network-warden/services/admins"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	"go.uber.org/fx"
)

var Dependencies = fx.Options(
	fx.Supply(toolkitfx.ServiceName(configurations.ServiceName)),
	fxlogger.Module,
	fxpostgres.Module,
	repository.Module,
	fx.Provide(
		grpc.NewHandler,
		grpc.NewGRPCServer,
		grpc.NewGatewayHandler,
		grpc.NewLivenessGateway,
		admins.New,
	),
)

var Invokes = fx.Invoke(
	fxgrpc.RunRegisteredGRPCServer,
	grpc.RunHTTPGateway,
	fxgrpc.RunHealthServer,
	fxgrpc.RunLivenessGateway,
)
