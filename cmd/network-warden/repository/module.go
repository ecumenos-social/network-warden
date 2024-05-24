package repository

import (
	holdersessions "github.com/ecumenos-social/network-warden/services/holder-sessions"
	"github.com/ecumenos-social/network-warden/services/holders"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		New,
		func(r *Repository) holders.Repository { return holders.Repository(r) },
		func(r *Repository) holdersessions.Repository { return holdersessions.Repository(r) },
	),
)
