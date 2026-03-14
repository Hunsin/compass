package flags

import "github.com/urfave/cli/v3"

// Shared flags
var (
	PostgresURL = cli.StringFlag{
		Name:    "postgres-url",
		Sources: cli.EnvVars("POSTGRES_URL"),
	}

	RedisURL = cli.StringFlag{
		Name:    "redis-url",
		Sources: cli.EnvVars("REDIS_URL"),
	}

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

	PartitionTable = cli.StringFlag{
		Name:  "table",
		Usage: "Specific table to create partition for (e.g. ohlcv_per_min, ohlcv_per_30min, ohlcv_per_day)",
	}
)
