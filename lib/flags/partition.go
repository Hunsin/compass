package flags

import "github.com/urfave/cli/v3"

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
