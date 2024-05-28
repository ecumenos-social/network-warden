package main

import (
	"github.com/ecumenos-social/network-warden/cmd/admin/configurations"
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
