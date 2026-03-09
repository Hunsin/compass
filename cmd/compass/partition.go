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
		Usage: "Create partitions for OHLCV tables",
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
	if year != 0 && month != 0 {
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

	queries := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ohlcv_per_min_%s PARTITION OF ohlcv_per_min FOR VALUES FROM ('%s') TO ('%s');`, monthStr, dateStart, dateEnd),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ohlcv_per_30min_%s PARTITION OF ohlcv_per_30min FOR VALUES FROM ('%s') TO ('%s');`, monthStr, dateStart, dateEnd),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS ohlcv_per_day_%s PARTITION OF ohlcv_per_day FOR VALUES FROM ('%s') TO ('%s');`, monthStr, dateStartDay, dateEndDay),
	}

	for _, query := range queries {
		log.Printf("Executing: %s", query)
		if _, err := pool.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}
