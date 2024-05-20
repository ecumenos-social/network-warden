package fxlogger

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/ecumenos-social/network-warden/pkg/toolkitfx"
)

type Config struct {
	Production bool `default:"true"`
}

var Module = fx.Options(
	fx.Provide(
		func(lc fx.Lifecycle, serviceName toolkitfx.ServiceName, cfg Config) (*zap.Logger, error) {
			var (
				logger *zap.Logger
				err    error
			)
			if cfg.Production {
				logger, err = NewProductionLogger(string(serviceName))
			} else {
				logger, err = NewDevelopmentLogger(string(serviceName))
			}
			if err != nil {
				return nil, err
			}
			zap.ReplaceGlobals(logger)

			lc.Append(fx.Hook{
				OnStart: nil,
				OnStop: func(ctx context.Context) error {
					_ = logger.Sync()
					return nil
				},
			})

			return logger, nil
		},
	),
)
