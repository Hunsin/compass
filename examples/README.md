# Examples

Python scripts that demonstrate how to interact with the Compass services.

## Prerequisites

Make sure the Compass services are running before executing the examples:

```sh
make start          # Start third-party services (Postgres, Redis, Keycloak)
make start-api      # Start the API service
make start-quote    # Start the Quote gRPC service
make start-statistics  # Start the Statistics gRPC service
```

## Setting Up a Python Environment

The easiest way to run the example scripts is inside a Docker container that
shares the same network as the Compass services.

### 1. Start a Python container

```sh
docker run -it \
  -v .:/compass \
  -w /compass/examples \
  --network compass \
  python:3.12 bash
```

> [!NOTE]
> The `--network compass` flag connects the container to the Docker network
> created by `docker compose`, allowing it to resolve service hostnames such as
> `compass-api-service` and `compass-quote-service`.

### 2. Create a virtual environment and install dependencies

```sh
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
```

## Running the Examples

When running inside the Docker container, use the **container names** as
hostnames instead of `localhost`.

### API Examples

```sh
# Log in via Keycloak and print the tokens
python api/login.py --server http://compass-api-service:50189

# Fetch current user info (requires access token from login)
python api/me.py --server http://compass-api-service:50189 --token <ACCESS_TOKEN>
```

### Quote Examples

```sh
# Create exchanges and securities
python quote/create_securities.py --server compass-quote-service:50168

# Upload historical quotes from CSV
python quote/csv_upload.py --server compass-quote-service:50168

# Get historical quotes
python quote/get_history.py --server compass-quote-service:50168
```

### Statistics Examples

```sh
# Upload margin transactions
python statistics/upload_margin_transactions.py --server compass-statistics-service:50178 --token <ACCESS_TOKEN>

# Get margin transactions
python statistics/get_margin_transactions.py --server compass-statistics-service:50178 --token <ACCESS_TOKEN>
```

> [!TIP]
> The Statistics service requires authentication. Use the access token obtained
> from `api/login.py`.
