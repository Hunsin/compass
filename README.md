# Compass

## Prerequisites

- [Docker Desktop](https://docs.docker.com/get-started/introduction/get-docker-desktop/) & Docker Compose
- git
- Make

## Development

### Common Commands

```sh
# Start third-party services
make start

# Stop services
make stop

# Stop services and remove associated volumes, images and networks
make clean

# Access PostgreSQL CLI
make psql

# Run migrations
make migrate-up      # Apply pending migrations
make migrate-down    # Rollback last migration
make migrate-version # Check migration version

# Generate Go code from SQL queries
make sqlc

# Generate Go and Python code from Protobuf definitions
make proto

# Generate mock implementations
make mock

# Run tests
make test     # Unit tests
make test-all # All tests including integration tests
```

### Project Structure

```text
.
├── cmd/                 # Application entry point
├── dockerfiles/         # Docker images
├── lib/                 # Domain model layer
├── postgres/
│   ├── gen/             # Generated Go code from SQL queries
│   ├── migrations/      # Database migrations
│   └── queries/         # SQL queries for sqlc
├── protocols/
│   ├── gen/             # Generated Go and Python code from Protobuf definitions
│   └── quote/           # Protobuf definitions for the quote service
└── services/            # gRPC service controllers

```

**Note:** Do not commit generated code. The outputs of `make sqlc`, `make mock`, and `make proto` are regenerated as needed and should not be tracked in version control.

### Testing

Before running tests, make sure the generated code and mocks are up to date:

```sh
make sqlc
make proto
make mock # depends on sqlc, must run after it
```

Tests are run inside the dev container:

```sh
make test
```

To run integration tests, make sure the third-party services are running (see `make start`):

```sh
make test-all
```

## Database

### Migrations

Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate).

### SQL Code Generation

The project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries. Configuration is defined in `sqlc.yaml`.

## Protobuf

Protobuf definitions live in `protocols/`. Code generation is handled by [buf](https://buf.build/) using remote plugins — no local plugin installation required. Configuration is in `buf.yaml` and `buf.gen.yaml`.
