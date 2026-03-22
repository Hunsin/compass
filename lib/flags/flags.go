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

	GRPCAddr = cli.StringFlag{
		Name:    "grpc-addr",
		Usage:   "gRPC server listen address",
		Sources: cli.EnvVars("GRPC_ADDR"),
	}

	HTTPAddr = cli.StringFlag{
		Name:    "http-addr",
		Usage:   "HTTP gateway listen address (grpc-gateway JSON)",
		Sources: cli.EnvVars("HTTP_ADDR"),
	}

	KeycloakURL = cli.StringFlag{
		Name:    "keycloak-url",
		Usage:   "Base URL of Keycloak server",
		Sources: cli.EnvVars("KEYCLOAK_URL"),
	}

	KeycloakRealm = cli.StringFlag{
		Name:    "keycloak-realm",
		Usage:   "Keycloak realm name",
		Sources: cli.EnvVars("KEYCLOAK_REALM"),
	}

	KeycloakClientID = cli.StringFlag{
		Name:    "keycloak-client-id",
		Usage:   "Keycloak client ID",
		Sources: cli.EnvVars("KEYCLOAK_CLIENT_ID"),
	}

	KeycloakClientSecret = cli.StringFlag{
		Name:    "keycloak-client-secret",
		Usage:   "Keycloak client secret (optional for public clients)",
		Sources: cli.EnvVars("KEYCLOAK_CLIENT_SECRET"),
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
