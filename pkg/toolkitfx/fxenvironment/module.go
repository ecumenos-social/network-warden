package fxenvironment

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/fx"

	"github.com/ecumenos-social/network-warden/pkg/toolkitfx"
)

var Module = func(cfg interface{}, useServicePrefix bool) fx.Option {
	return fx.Options(
		fx.Provide(NewEnvConfig(cfg, useServicePrefix)),
		fx.Invoke(
			RaiseEnvConfigError,
		),
	)
}

type FxConfig interface{}

type ConfigError struct {
	value error
}

func NewEnvConfig(cfg interface{}, useServicePrefix bool) func(sn toolkitfx.ServiceName) (FxConfig, ConfigError) {
	return func(sn toolkitfx.ServiceName) (FxConfig, ConfigError) {
		prefix := ""
		if useServicePrefix {
			prefix = string(sn)
		}
		if err := envconfig.Process(prefix, cfg); err != nil {
			return cfg, ConfigError{value: fmt.Errorf("error: parsing config: %s", err)}
		}
		return cfg, ConfigError{value: nil}
	}
}

func RaiseEnvConfigError(ce ConfigError) error {
	return ce.value
}
