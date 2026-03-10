package flags

import "github.com/urfave/cli/v3"

var PostgresURL = cli.StringFlag{
	Name:    "postgres-url",
	Sources: cli.EnvVars("POSTGRES_URL"),
}
