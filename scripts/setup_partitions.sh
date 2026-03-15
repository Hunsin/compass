#!/bin/bash
set -e

echo "Creating default partitions for OHLCV tables..."

docker exec -i compass-postgres psql -U compass -d compass <<EOF
CREATE TABLE IF NOT EXISTS ohlcv_per_min_default PARTITION OF ohlcv_per_min DEFAULT;
CREATE TABLE IF NOT EXISTS ohlcv_per_30min_default PARTITION OF ohlcv_per_30min DEFAULT;
CREATE TABLE IF NOT EXISTS ohlcv_per_day_default PARTITION OF ohlcv_per_day DEFAULT;
EOF

echo "Default partitions created successfully."
