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
		Usage: "Create partitions for OHLCV tables (default: creates partitions for current and next month)",
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

func createPartitions(ctx context.Context, pool *pgxpool.Pool, target time.Time) error {
	nextMonth := target.AddDate(0, 1, 0)

	monthStr := target.Format("2006_01")
	dateStart := target.Format("2006-01-02 00:00:00")
	dateEnd := nextMonth.Format("2006-01-02 00:00:00")
	dateStartDay := target.Format("2006-01-02")
	dateEndDay := nextMonth.Format("2006-01-02")

	queries := []struct {
		table string
		query string
	}{
		{
			"ohlcv_per_min",
			fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ohlcv_per_min_%s PARTITION OF ohlcv_per_min FOR VALUES FROM ('%s') TO ('%s');`,
				monthStr, dateStart, dateEnd),
		},
		{
			"ohlcv_per_30min",
			fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ohlcv_per_30min_%s PARTITION OF ohlcv_per_30min FOR VALUES FROM ('%s') TO ('%s');`,
				monthStr, dateStart, dateEnd),
		},
		{
			"ohlcv_per_day",
			fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ohlcv_per_day_%s PARTITION OF ohlcv_per_day FOR VALUES FROM ('%s') TO ('%s');`,
				monthStr, dateStartDay, dateEndDay),
		},
	}

	for _, item := range queries {
		log.Printf("Executing: %s", item.query)

		execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		_, err := pool.Exec(execCtx, item.query)
		cancel()

		if err != nil {
			return fmt.Errorf("failed to create partition for table %s (month %s): %w", item.table, monthStr, err)
		}
	}

	return nil
}
