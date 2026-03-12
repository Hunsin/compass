package main

import (
	"context"
	"os"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

func main() {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cmd := &cli.Command{
		Name:     "compass",
		Commands: []*cli.Command{quoteCommand()},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal().Err(err).Msg("failed running compass")
	}
}
