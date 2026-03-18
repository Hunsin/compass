package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/urfave/cli/v3"

	"github.com/Hunsin/compass/lib/flags"
	"github.com/Hunsin/compass/lib/logutil"
)

func partitionCommand() *cli.Command {
	return &cli.Command{
		Name:  "partition",
		Usage: "Create partitions for OHLCV tables (min/30min per month, day per year)",
		Flags: []cli.Flag{
			&flags.PostgresURL,
			&flags.PartitionYear,
			&flags.PartitionMonth,
			&flags.PartitionTable,
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

	var (
		year        = int(cmd.Int(flags.PartitionYear.Name))
		month       = int(cmd.Int(flags.PartitionMonth.Name))
		targetTable = cmd.String(flags.PartitionTable.Name)
	)

	log := logutil.DefaultLogger(os.Stdout)
	ctx = logutil.WithLogger(ctx, log)
	switch targetTable {
	case ohlcvDayTableName:
		// ohlcv_per_day uses yearly partitions
		var targets targetTimes
		if year != 0 {
			if err := targets.addYear(year); err != nil {
				return err
			}
		} else {
			targets.addDefaultYears()
		}
		for _, t := range targets {
			if err := createYearPartition(ctx, pool, t, ohlcvDayTableName); err != nil {
				return fmt.Errorf("failed to create partition for %s: %w", t.Format("2006"), err)
			}
		}

	case ohlcvMinTableName, ohlcv30MinTableName:
		// Specific min/30min table uses monthly partitions
		var targets targetTimes
		if year != 0 || month != 0 {
			if err := targets.addMonth(year, month); err != nil {
				return err
			}
		} else {
			targets.addDefaultMonths()
		}
		for _, t := range targets {
			if err := createMonthPartition(ctx, pool, t, targetTable); err != nil {
				return fmt.Errorf("failed to create partition for %s: %w", t.Format("2006-01"), err)
			}
		}

	default:
		// No table specified: create partitions for all tables using their respective defaults
		var monthTargets targetTimes
		if year != 0 || month != 0 {
			if err := monthTargets.addMonth(year, month); err != nil {
				return err
			}
		} else {
			monthTargets.addDefaultMonths()
		}
		for _, t := range monthTargets {
			for _, table := range []string{ohlcvMinTableName, ohlcv30MinTableName} {
				if err := createMonthPartition(ctx, pool, t, table); err != nil {
					return fmt.Errorf("failed to create partition for %s: %w", t.Format("2006-01"), err)
				}
			}
		}

		var yearTargets targetTimes
		if year != 0 {
			if err := yearTargets.addYear(year); err != nil {
				return err
			}
		} else {
			yearTargets.addDefaultYears()
		}
		for _, t := range yearTargets {
			if err := createYearPartition(ctx, pool, t, ohlcvDayTableName); err != nil {
				return fmt.Errorf("failed to create partition for %s: %w", t.Format("2006"), err)
			}
		}
	}

	log.Info().Msg("Partitions created successfully.")
	return nil
}

// targetTimes is a list of target times for partition creation.
type targetTimes []time.Time

func (t *targetTimes) addMonth(year int, month int) error {
	if err := validateMonth(month); err != nil {
		return err
	}
	if err := validateYear(year); err != nil {
		return err
	}
	targetMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	*t = append(*t, targetMonth)
	return nil
}

func (t *targetTimes) addYear(year int) error {
	if err := validateYear(year); err != nil {
		return err
	}
	targetYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	*t = append(*t, targetYear)
	return nil
}

func (t *targetTimes) addDefaultMonths() {
	now := time.Now().UTC()
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	nextMonth := currentMonth.AddDate(0, 1, 0)
	*t = append(*t, currentMonth, nextMonth)
}

func (t *targetTimes) addDefaultYears() {
	now := time.Now().UTC()
	currentYear := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	nextYear := currentYear.AddDate(1, 0, 0)
	*t = append(*t, currentYear, nextYear)
}

func validateYear(year int) error {
	if year < 1911 || year > 2100 {
		return fmt.Errorf("invalid year %d: must be between 1911 and 2100", year)
	}
	return nil
}

func validateMonth(month int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month %d: must be between 1 and 12", month)
	}
	return nil
}

const (
	ohlcvMinTableName   = "ohlcv_per_min"
	ohlcv30MinTableName = "ohlcv_per_30min"
	ohlcvDayTableName   = "ohlcv_per_day"

	sqlCreatePartitionMin   = "CREATE TABLE IF NOT EXISTS ohlcv_per_min_%s PARTITION OF ohlcv_per_min FOR VALUES FROM ('%s') TO ('%s');"
	sqlCreatePartition30Min = "CREATE TABLE IF NOT EXISTS ohlcv_per_30min_%s PARTITION OF ohlcv_per_30min FOR VALUES FROM ('%s') TO ('%s');"
	sqlCreatePartitionDay   = "CREATE TABLE IF NOT EXISTS ohlcv_per_day_%s PARTITION OF ohlcv_per_day FOR VALUES FROM ('%s') TO ('%s');"
)

// createMonthPartition creates a monthly partition for the given table (ohlcv_per_min or ohlcv_per_30min).
// baseMonth should be the first day of the target month.
func createMonthPartition(ctx context.Context, pool *pgxpool.Pool, baseMonth time.Time, table string) error {
	nextMonth := baseMonth.AddDate(0, 1, 0)
	suffix := baseMonth.Format("2006_01")
	start := baseMonth.Format("2006-01-02 00:00:00")
	end := nextMonth.Format("2006-01-02 00:00:00")

	var query string
	switch table {
	case ohlcvMinTableName:
		query = fmt.Sprintf(sqlCreatePartitionMin, suffix, start, end)
	case ohlcv30MinTableName:
		query = fmt.Sprintf(sqlCreatePartition30Min, suffix, start, end)
	default:
		return fmt.Errorf("unsupported table for monthly partition: %s", table)
	}

	return execQuery(ctx, pool, query, table, suffix)
}

// createYearPartition creates a yearly partition for ohlcv_per_day.
// baseYear should be January 1st of the target year.
func createYearPartition(ctx context.Context, pool *pgxpool.Pool, baseYear time.Time, table string) error {
	nextYear := baseYear.AddDate(1, 0, 0)
	suffix := baseYear.Format("2006")
	start := baseYear.Format("2006-01-02")
	end := nextYear.Format("2006-01-02")

	var query string
	switch table {
	case ohlcvDayTableName:
		query = fmt.Sprintf(sqlCreatePartitionDay, suffix, start, end)
	default:
		return fmt.Errorf("unsupported table for yearly partition: %s", table)
	}

	return execQuery(ctx, pool, query, table, suffix)
}

func execQuery(ctx context.Context, pool *pgxpool.Pool, query, table, period string) error {
	log, ok := logutil.FromContext(ctx)
	if ok {
		log.Printf("Executing: %s", query)
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err := pool.Exec(execCtx, query)
	if err != nil {
		return fmt.Errorf("failed to create partition for table %s (%s): %w", table, period, err)
	}
	return nil
}
