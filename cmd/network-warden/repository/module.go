package repository

import (
	"github.com/ecumenos-social/network-warden/services/holders"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		New,
		func(r *Repository) holders.Repository {
			return holders.Repository(r)
		},
	),
)
