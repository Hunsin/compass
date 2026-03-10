package flags

import "github.com/urfave/cli/v3"

var ListenAddr = cli.StringFlag{
	Name:    "listen-addr",
	Sources: cli.EnvVars("LISTEN_ADDR"),
}
