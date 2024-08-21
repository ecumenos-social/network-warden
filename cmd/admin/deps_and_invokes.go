package main

import (
	"github.com/ecumenos-social/network-warden/cmd/admin/configurations"
	"github.com/ecumenos-social/network-warden/cmd/admin/grpc"
	"github.com/ecumenos-social/network-warden/cmd/admin/pgseeds"
	"github.com/ecumenos-social/network-warden/repository"
	"github.com/ecumenos-social/network-warden/services/adminauth"
	"github.com/ecumenos-social/network-warden/services/admins"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/network-warden/services/jwt"
	networknodes "github.com/ecumenos-social/network-warden/services/network-nodes"
	networkwardens "github.com/ecumenos-social/network-warden/services/network-wardens"
	personaldatanodes "github.com/ecumenos-social/network-warden/services/personal-data-nodes"
	"github.com/ecumenos-social/toolkitfx"
	"github.com/ecumenos-social/toolkitfx/fxgrpc"
	"github.com/ecumenos-social/toolkitfx/fxlogger"
	"github.com/ecumenos-social/toolkitfx/fxpostgres"
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
		adminauth.New,
		jwt.New,
		personaldatanodes.New,
		networkwardens.New,
		networknodes.New,
		idgenerators.NewAdminsIDGenerator,
		idgenerators.NewAdminSessionsIDGenerator,
		idgenerators.NewPersonalDataNodesIDGenerator,
		idgenerators.NewNetworkNodesIDGenerator,
		idgenerators.NewNetworkWardensIDGenerator,
		pgseeds.New,
	),
)

var Invokes = fx.Invoke(
	fxgrpc.RunRegisteredGRPCServer,
	grpc.RunHTTPGateway,
	fxgrpc.RunHealthServer,
	fxgrpc.RunLivenessGateway,
)
