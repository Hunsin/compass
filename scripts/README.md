# Scripts

This directory contains utility scripts used for local development, database initialization, and deployment.

## Files

### `compass-realm.json`

Keycloak realm import configuration used to initialize the local Keycloak instance.

When Keycloak starts with the `--import-realm` flag (see `compose.yaml`), it reads this file from the mounted path `/opt/keycloak/data/import/compass-realm.json` and automatically provisions:

- **Realm**: `compass-realm`
- **Client**: `compass-backend` (public OIDC client with direct access grants enabled)
- **Test user**: `testuser` / `123456` (email: `testuser@example.com`)

This is intended for **local development only**. Do not use these credentials in production.

---

### `create-databases.sh`

A PostgreSQL [initialization script](https://hub.docker.com/_/postgres#:~:text=Initialization%20scripts) mounted into the Postgres container at `/docker-entrypoint-initdb.d/`. When the container's data directory is first created, this script runs automatically and:

1. Creates the Keycloak database user (`$KEYCLOAK_DB_USER`) with the configured password.
2. Creates the Keycloak database (`$KEYCLOAK_DB`).
3. Grants all necessary privileges to the Keycloak user on the database and `public` schema.

The required environment variables (`KEYCLOAK_DB_USER`, `KEYCLOAK_DB_PASSWORD`, `KEYCLOAK_DB`) are supplied from the `.env` file via `compose.yaml`.

> [!WARNING]
> **Known Limitation**: Scripts in `/docker-entrypoint-initdb.d/` only execute when the PostgreSQL data directory is initialized for the **first time**. If you remove the container image (or the volume) and rebuild, the initialization may not re-run as expected depending on volume state. Essentially, deleting the image alone is not enough to trigger re-initialization — you need to also remove the associated volume (`docker volume rm ...` or `make clean`).

<!-- TODO: Redesign the Keycloak database initialization approach to be idempotent
     and not rely solely on the docker-entrypoint-initdb.d mechanism, which only
     runs on first-time volume creation. Consider using a dedicated init container
     or a migration-based approach. -->

---

### `migrate-entrypoint.sh`

> **Status: Deprecated — No longer in use.**

This script was created as a custom entrypoint for the `migrate` container (see `dockerfiles/migrate.Dockerfile`) to support AWS RDS IAM authentication.

**Background**: An earlier AWS RDS (Aurora) deployment was set up using the "Easy create" configuration, which only supported IAM-based token authentication for database access. This script would:

1. Generate a temporary IAM auth token via `aws rds generate-db-auth-token`.
2. URL-encode the token (since it contains special characters).
3. Construct a `DATABASE_URL` with the token as the password.
4. Execute the `migrate` command with that URL.

**Why it's no longer used**: Keycloak does not support IAM token-based database authentication, so the RDS instance was later recreated with standard username/password authentication configured manually. The local `compose.yaml` now uses the stock `migrate/migrate` image directly (without this entrypoint).

This file is kept for reference. It can be safely removed if no longer needed.

---

### `setup_partitions.sh`

Creates default partitions for the OHLCV (Open/High/Low/Close/Volume) partitioned tables in the local PostgreSQL database:

- `ohlcv_per_min_default`
- `ohlcv_per_30min_default`
- `ohlcv_per_day_default`

**Usage**:

```bash
# Directly
bash scripts/setup_partitions.sh

# Via Makefile
make partition
```

This script connects to the `compass-postgres` container via `docker exec` and runs the partition DDL.

> [!IMPORTANT]
> This must be run **before** integration tests (`make test-all`), which depends on the `partition` target. Without these default partitions, tests involving OHLCV data will fail.

This script is for **local development/testing only** and is not used in production environments.
