package main

import (
	"github.com/ecumenos-social/network-warden/cmd/network-warden/configurations"
	"github.com/ecumenos-social/network-warden/cmd/network-warden/grpc"
	"github.com/ecumenos-social/network-warden/cmd/network-warden/pgseeds"
	"github.com/ecumenos-social/network-warden/cmd/network-warden/repository"
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	"github.com/ecumenos-social/network-warden/services/holders"
	"github.com/ecumenos-social/network-warden/services/idgenerators"
	"github.com/ecumenos-social/network-warden/services/jwt"
	networknodes "github.com/ecumenos-social/network-warden/services/network-nodes"
	networkwardens "github.com/ecumenos-social/network-warden/services/network-wardens"
	personaldatanodes "github.com/ecumenos-social/network-warden/services/personal-data-nodes"
	smssender "github.com/ecumenos-social/network-warden/services/sms-sender"
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
		grpc.NewGRPCServer,
		grpc.NewHTTPGatewayHandler,
		grpc.NewLivenessGateway,
		grpc.NewHandler,
		holders.New,
		auth.New,
		jwt.New,
		emailer.New,
		smssender.New,
		networknodes.New,
		personaldatanodes.New,
		networkwardens.New,
		idgenerators.NewHolderSessionsIDGenerator,
		idgenerators.NewHoldersIDGenerator,
		idgenerators.NewNetworkNodesIDGenerator,
		idgenerators.NewPersonalDataNodesIDGenerator,
		idgenerators.NewNetworkWardensIDGenerator,
		idgenerators.NewSentEmailsIDGenerator,
		pgseeds.New,
	),
)

var Invokes = fx.Invoke(
	fxgrpc.RunRegisteredGRPCServer,
	grpc.RunHTTPGateway,
	fxgrpc.RunHealthServer,
	fxgrpc.RunLivenessGateway,
)
