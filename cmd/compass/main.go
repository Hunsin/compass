package main

import (
	"context"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
)

// defaultLogger creates a zerolog logger that writes to w with timestamps enabled.
func defaultLogger(w io.Writer) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().Logger()
}

func main() {
	cmd := &cli.Command{
		Name: "compass",
		Commands: []*cli.Command{
			apiCommand(),
			partitionCommand(),
			quoteCommand(),
		},
	}

	// Set up the default logger
	log := defaultLogger(os.Stdout)
	zerolog.DefaultContextLogger = &log

	ctx := log.WithContext(context.Background())
	if err := cmd.Run(ctx, os.Args); err != nil {
		l := log.Output(os.Stderr)
		l.Fatal().Err(err).Msg("failed running compass")
	}
}
