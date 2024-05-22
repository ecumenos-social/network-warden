package main

import (
	"github.com/ecumenos-social/network-warden/cmd/network-warden/configurations"
	"github.com/ecumenos-social/network-warden/pkg/fxpostgres"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/fx"
)

var runAppCmd = &cli.Command{
	Name:  "run",
	Usage: "running server",
	Flags: configurations.Flags,
	Action: func(cctx *cli.Context) error {
		fx.New(
			configurations.Module(cctx),
			Dependencies,
			Invokes,
			fx.StartTimeout(configurations.StartTimeout),
		).Run()

		return nil
	},
}

var runMigrateUpCmd = &cli.Command{
	Name:  "migrate-up",
	Usage: "migrate database(s) up",
	Flags: configurations.Flags,
	Action: func(cctx *cli.Context) error {
		fx.New(
			configurations.Module(cctx),
			Dependencies,
			Invokes,
			fx.StartTimeout(configurations.StartTimeout),
			fx.Invoke(func(runner *fxpostgres.MigrationsRunner) error {
				return runner.MigrateUp()
			}),
		).Run()

		return nil
	},
}

var runMigrateDownCmd = &cli.Command{
	Name:  "migrate-down",
	Usage: "migrate database(s) down",
	Flags: configurations.Flags,
	Action: func(cctx *cli.Context) error {
		fx.New(
			configurations.Module(cctx),
			Dependencies,
			Invokes,
			fx.StartTimeout(configurations.StartTimeout),
			fx.Invoke(func(runner *fxpostgres.MigrationsRunner) error {
				return runner.MigrateDown()
			}),
		).Run()

		return nil
	},
}
