package repository

import (
	"github.com/ecumenos-social/network-warden/services/adminauth"
	"github.com/ecumenos-social/network-warden/services/admins"
	"github.com/ecumenos-social/network-warden/services/auth"
	"github.com/ecumenos-social/network-warden/services/emailer"
	"github.com/ecumenos-social/network-warden/services/holders"
	networknodes "github.com/ecumenos-social/network-warden/services/network-nodes"
	networkwardens "github.com/ecumenos-social/network-warden/services/network-wardens"
	personaldatanodes "github.com/ecumenos-social/network-warden/services/personal-data-nodes"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		New,
		func(r *Repository) holders.Repository { return holders.Repository(r) },
		func(r *Repository) auth.Repository { return auth.Repository(r) },
		func(r *Repository) emailer.Repository { return emailer.Repository(r) },
		func(r *Repository) networknodes.Repository { return networknodes.Repository(r) },
		func(r *Repository) personaldatanodes.Repository { return personaldatanodes.Repository(r) },
		func(r *Repository) networkwardens.Repository { return networkwardens.Repository(r) },
		func(r *Repository) admins.Repository { return admins.Repository(r) },
		func(r *Repository) adminauth.Repository { return adminauth.Repository(r) },
	),
)
