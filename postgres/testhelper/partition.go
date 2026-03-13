// Package testhelper provides utilities for setting up the database in integration tests.
package testhelper

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// defaultPartitionStmts are DDL statements that create DEFAULT partitions for
// all OHLCV partition tables. They accept any rows that fall outside the range
// of explicitly defined partitions, preventing insert failures during tests.
var defaultPartitionStmts = []string{
	`CREATE TABLE IF NOT EXISTS ohlcv_per_min_default PARTITION OF ohlcv_per_min DEFAULT`,
	`CREATE TABLE IF NOT EXISTS ohlcv_per_30min_default PARTITION OF ohlcv_per_30min DEFAULT`,
	`CREATE TABLE IF NOT EXISTS ohlcv_per_day_default PARTITION OF ohlcv_per_day DEFAULT`,
}

// CreateDefaultPartitions creates DEFAULT partitions for all OHLCV partition
// tables. It is intended for use in integration tests only — production
// deployments should create explicit range partitions instead.
func CreateDefaultPartitions(ctx context.Context, pool *pgxpool.Pool) error {
	for _, stmt := range defaultPartitionStmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("testhelper: create default partition with statement %q: %w", stmt, err)
		}
	}
	return nil
}
