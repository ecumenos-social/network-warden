package repository

import (
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	"github.com/ecumenos-social/network-warden/services/holders"
	networknodes "github.com/ecumenos-social/network-warden/services/network-nodes"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		New,
		func(r *Repository) holders.Repository { return holders.Repository(r) },
		func(r *Repository) auth.Repository { return auth.Repository(r) },
		func(r *Repository) emailer.Repository { return emailer.Repository(r) },
		func(r *Repository) networknodes.Repository { return networknodes.Repository(r) },
	),
)
