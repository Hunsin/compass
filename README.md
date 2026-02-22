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

# Stop services and remove images
make stop

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
```

## Project Structure

```
.
├── dockerfiles/         # Docker images
├── postgres/
│   ├── gen/             # Generated Go code from SQL queries
│   ├── migrations/      # Database migrations
│   └── queries/         # SQL queries for sqlc
├── protocols/           # Protobuf definitions
└── lib/                 # Go libraries
```

## Database

### Migrations

Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate).

### SQL Code Generation

The project uses [sqlc](https://sqlc.dev/) to generate type-safe Go code from SQL queries. Configuration is defined in `sqlc.yaml`.
