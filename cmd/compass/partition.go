package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	"github.com/Hunsin/compass/lib/flags"
)

func partitionCommand() *cli.Command {
	return &cli.Command{
		Name:  "partition",
		Usage: "Create partitions for OHLCV tables (min/30min per month, day per year)",
		Flags: []cli.Flag{
			&flags.PostgresURL,
			&flags.PartitionYear,
			&flags.PartitionMonth,
		},
		Action: partitionAction,
	}
}

func partitionAction(ctx context.Context, cmd *cli.Command) error {
	pool, err := pgxpool.New(ctx, cmd.String(flags.PostgresURL.Name))
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	childCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(childCtx); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	year := int(cmd.Int(flags.PartitionYear.Name))
	month := int(cmd.Int(flags.PartitionMonth.Name))

	var targets []time.Time
	// Create partitions for the specified month, or the current and next month if not specified
	if year != 0 || month != 0 {
		if year == 0 || month == 0 {
			return fmt.Errorf("both year and month must be provided together, or neither")
		}
		if month < 1 || month > 12 {
			return fmt.Errorf("invalid month %d: must be between 1 and 12", month)
		}
		if year < 1911 || year > 2100 {
			return fmt.Errorf("invalid year %d: must be between 1911 and 2100", year)
		}
		targets = append(targets, time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC))
	} else {
		now := time.Now().UTC()
		currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		nextMonth := currentMonth.AddDate(0, 1, 0)
		targets = append(targets, currentMonth, nextMonth)
	}

	for _, t := range targets {
		if err := createPartitions(ctx, pool, t); err != nil {
			return fmt.Errorf("failed to create partition for %s: %w", t.Format("2006-01"), err)
		}
	}

	log.Println("Partitions created successfully.")
	return nil
}

const (
	sqlCreatePartitionMin   = "CREATE TABLE IF NOT EXISTS ohlcv_per_min_%s PARTITION OF ohlcv_per_min FOR VALUES FROM ('%s') TO ('%s');"
	sqlCreatePartition30Min = "CREATE TABLE IF NOT EXISTS ohlcv_per_30min_%s PARTITION OF ohlcv_per_30min FOR VALUES FROM ('%s') TO ('%s');"
	sqlCreatePartitionDay   = "CREATE TABLE IF NOT EXISTS ohlcv_per_day_%s PARTITION OF ohlcv_per_day FOR VALUES FROM ('%s') TO ('%s');"
)

func createPartitions(ctx context.Context, pool *pgxpool.Pool, baseMonth time.Time) error {
	nextMonth := baseMonth.AddDate(0, 1, 0)

	// For monthly partitions (used by ohlcv_per_min and ohlcv_per_30min)
	monthSuffix := baseMonth.Format("2006_01")
	timeBoundStart := baseMonth.Format("2006-01-02 00:00:00")
	timeBoundEnd := nextMonth.Format("2006-01-02 00:00:00")

	// For yearly partitions (used by ohlcv_per_day)
	baseYear := time.Date(baseMonth.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	nextYear := baseYear.AddDate(1, 0, 0)
	yearSuffix := baseYear.Format("2006")
	dateBoundStart := baseYear.Format("2006-01-02")
	dateBoundEnd := nextYear.Format("2006-01-02")

	queries := []struct {
		table string
		query string
	}{
		{
			"ohlcv_per_min",
			fmt.Sprintf(sqlCreatePartitionMin, monthSuffix, timeBoundStart, timeBoundEnd),
		},
		{
			"ohlcv_per_30min",
			fmt.Sprintf(sqlCreatePartition30Min, monthSuffix, timeBoundStart, timeBoundEnd),
		},
		{
			"ohlcv_per_day",
			fmt.Sprintf(sqlCreatePartitionDay, yearSuffix, dateBoundStart, dateBoundEnd),
		},
	}

	for _, item := range queries {
		log.Printf("Executing: %s", item.query)

		execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		_, err := pool.Exec(execCtx, item.query)
		if err != nil {
			return fmt.Errorf("failed to create partition for table %s (base month %s): %w", item.table, monthSuffix, err)
		}
	}

	return nil
}
