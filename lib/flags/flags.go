package flags

import "github.com/urfave/cli/v3"

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
