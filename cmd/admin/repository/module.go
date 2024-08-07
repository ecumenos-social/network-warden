package repository

import (
	"github.com/ecumenos-social/network-warden/services/admins"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		New,
		func(r *Repository) admins.Repository { return admins.Repository(r) },
	),
)
