#!/bin/sh
set -e

# Generate IAM auth token as the database password.
# Required environment variables:
#   DB_HOST     - Aurora cluster endpoint
#   DB_PORT     - Database port (default: 5432)
#   DB_USER     - Database user configured for IAM auth
#   DB_NAME     - Target database name
#   AWS_REGION  - AWS region of the Aurora cluster

: "${DB_HOST:?DB_HOST is required}"
: "${DB_USER:?DB_USER is required}"
: "${DB_NAME:?DB_NAME is required}"
: "${AWS_REGION:?AWS_REGION is required}"
DB_PORT="${DB_PORT:-5432}"

TOKEN=$(aws rds generate-db-auth-token \
  --hostname "$DB_HOST" \
  --port "$DB_PORT" \
  --username "$DB_USER" \
  --region "$AWS_REGION")

# URL-encode the token as it contains special characters (?, &, =)
TOKEN=$(printf '%s' "$TOKEN" | python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.stdin.read(), safe=''))")

DATABASE_URL="postgres://${DB_USER}:${TOKEN}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require"

exec migrate -path=/migrations -database="$DATABASE_URL" "$@"
