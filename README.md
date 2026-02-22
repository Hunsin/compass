# Compass

## Prerequisites

- [Docker Desktop](https://docs.docker.com/get-started/introduction/get-docker-desktop/) & Docker Compose
- git
- Make

## Quick Start

1. Start the services:
```bash
make start
```

2. Run database migrations:
```bash
make migrate-up
```

3. Generate Go code from SQL:
```bash
make sqlc
```

## Development

### Common Commands

```bash
# Start all services
make start

# Stop services
make stop

# Stop services and remove images
make clean

# Access development container
make go

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
```

## Project Structure

```
.
├── dockerfiles/         # Docker images
├── postgres/
│   ├── gen/             # Generated Go code from SQL queries
│   ├── migrations/      # Database migrations
│   └── queries/         # SQL queries for sqlc
└── protocols/
    ├── gen/
    │   ├── go/          # Generated Go code from Protobuf definitions
    │   └── python/      # Generated Python code from Protobuf definitions
    └── quote/           # Protobuf definitions for the quote service
```

## Database

### Migrations

Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate).

### SQL Code Generation

The project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries. Configuration is defined in `sqlc.yaml`.

## Protobuf

Protobuf definitions live in `protocols/`. Code generation is handled by [buf](https://buf.build/) using remote plugins — no local plugin installation required. Configuration is in `buf.yaml` and `buf.gen.yaml`.
