package postgres

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"

	// this impost is needed for running migration down
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// this impost is needed for running migration down
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewMigrateUpFunc() func(sourceURL, dbURL string, log *zap.Logger, shutdowner fx.Shutdowner) error {
	return func(sourceURL, dbURL string, log *zap.Logger, shutdowner fx.Shutdowner) error {
		log.Info("command starting...")
		defer func(l *zap.Logger) {
			l.Info("command finished")
		}(log)

		if err := runMigration(sourceURL, dbURL, func(m *migrate.Migrate) error { return m.Up() }); err != nil {
			return err
		}

		return shutdowner.Shutdown()

	}
}

func NewMigrateDownFunc() func(sourceURL, dbURL string, log *zap.Logger, shutdowner fx.Shutdowner) error {
	return func(sourceURL, dbURL string, log *zap.Logger, shutdowner fx.Shutdowner) error {
		log.Info("command starting...")
		defer func(l *zap.Logger) {
			l.Info("command finished")
		}(log)

		if err := runMigration(sourceURL, dbURL, func(m *migrate.Migrate) error { return m.Down() }); err != nil {
			return err
		}

		return shutdowner.Shutdown()

	}
}

func runMigration(sourceURL, dbURL string, cb func(m *migrate.Migrate) error) error {
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return err
	}
	if err = cb(m); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
