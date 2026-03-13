package flags

import "github.com/urfave/cli/v3"

// Shared flags
var (
	PostgresURL = cli.StringFlag{
		Name:    "postgres-url",
		Sources: cli.EnvVars("POSTGRES_URL"),
	}
)

// Quote flags
var (
	ListenAddr = cli.StringFlag{
		Name:    "listen-addr",
		Sources: cli.EnvVars("LISTEN_ADDR"),
	}
)

// Partition flags
var (
	PartitionYear = cli.IntFlag{
		Name:  "year",
		Usage: "Year to create partition for (e.g. 2026)",
	}

	PartitionMonth = cli.IntFlag{
		Name:  "month",
		Usage: "Month to create partition for (1-12)",
	}
)
