package fxpostgres

import (
	"fmt"
	"os"
	"strings"

	"github.com/ecumenos-social/network-warden/pkg/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type MigrationsRunner struct {
	prod           bool
	url            string
	migrationsPath string
	logger         *zap.Logger
	shutdowner     fx.Shutdowner
}

func NewMigrationsRunner(config *Config, logger *zap.Logger, shutdowner fx.Shutdowner) *MigrationsRunner {
	for _, env := range os.Environ() {
		// env is
		envPair := strings.SplitN(env, "=", 2)
		key := envPair[0]
		value := envPair[1]

		fmt.Printf("%s : %s\n", key, value)
	}
	fmt.Println("@!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!", config.URL, config.MigrationsPath)
	return &MigrationsRunner{
		url:            config.URL,
		migrationsPath: config.MigrationsPath,
		logger:         logger,
		shutdowner:     shutdowner,
	}
}

func (r *MigrationsRunner) MigrateUp() error {
	fn := postgres.NewMigrateUpFunc()
	if !r.prod {
		r.logger.Info("runnning migrate up",
			zap.String("db_url", r.url),
			zap.String("source_path", r.migrationsPath))
	}
	return fn(r.migrationsPath, r.url+"?sslmode=disable", r.logger, r.shutdowner)
}

func (r *MigrationsRunner) MigrateDown() error {
	fn := postgres.NewMigrateDownFunc()
	if !r.prod {
		r.logger.Info("runnning migrate down",
			zap.String("db_url", r.url),
			zap.String("source_path", r.migrationsPath))
	}
	return fn(r.migrationsPath, r.url+"?sslmode=disable", r.logger, r.shutdowner)
}
