package main

import (
	"fmt"
	"os"

	"github.com/ecumenos-social/network-warden/cmd/network-warden/configurations"
	"github.com/joho/godotenv"
	cli "github.com/urfave/cli/v2"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("error loading .env file", "err", err)
		os.Exit(-1)
	}

	if err := run(os.Args); err != nil {
		fmt.Println("exiting", "err", err)
		os.Exit(-1)
	}
}

func run(args []string) error {
	app := cli.App{
		Name:    configurations.ServiceName,
		Usage:   "managing networks service",
		Version: string(configurations.ServiceVersion),
		Flags:   []cli.Flag{},
		Commands: []*cli.Command{
			runAppCmd,
			runMigrateUpCmd,
			runMigrateDownCmd,
			runSeedCmd,
		},
	}

	return app.Run(args)
}
