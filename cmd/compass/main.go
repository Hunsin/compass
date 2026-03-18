package main

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"github.com/Hunsin/compass/lib/logutil"
)

func main() {
	log := logutil.DefaultLogger(os.Stderr)

	cmd := &cli.Command{
		Name: "compass",
		Commands: []*cli.Command{
			quoteCommand(),
			partitionCommand(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err).Msg("failed running compass")
	}
}
