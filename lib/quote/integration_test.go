//go:build integration

package quote

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/lib/quote/testdata"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

func connectPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("POSTGRES_URL")
	if url == "" {
		t.Skip("POSTGRES_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}

func connectRedis(t *testing.T) *redis.Client {
	t.Helper()
	url := os.Getenv("REDIS_URL")
	if url == "" {
		t.Skip("REDIS_URL not set")
	}
	opts, err := redis.ParseURL(url)
	require.NoError(t, err)
	rdb := redis.NewClient(opts)
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func assertBucket(t *testing.T, want, got *pb.OHLCV) {
	t.Helper()
	require.Equal(t, want.GetTs().AsTime(), got.GetTs().AsTime(), "ts")
	require.InDelta(t, want.GetOpen(), got.GetOpen(), 0.01, "open at %v", want.GetTs())
	require.InDelta(t, want.GetHigh(), got.GetHigh(), 0.01, "high at %v", want.GetTs())
	require.InDelta(t, want.GetLow(), got.GetLow(), 0.01, "low at %v", want.GetTs())
	require.InDelta(t, want.GetClose(), got.GetClose(), 0.01, "close at %v", want.GetTs())
	require.Equal(t, want.GetVolume(), got.GetVolume(), "volume at %v", want.GetTs())
}

func sp(s string) *string { return &s }

func TestCreateOHLCVsPerMin(t *testing.T) {
	const exc, sym = "intg", "2454"
	var (
		pool = connectPool(t)
		rdb  = connectRedis(t)
		ctx  = context.Background()
		mdl  = Connect(pool, rdb)
	)

	// Ensure clean state before and after.
	pool.Exec(ctx, "DELETE FROM exchanges WHERE abbr = $1", exc)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM exchanges WHERE abbr = $1", exc) })

	require.NoError(t, mdl.CreateExchange(ctx, &pb.Exchange{
		Abbr: sp(exc), Name: sp("Integration Test"), Timezone: sp("UTC"),
	}))
	require.NoError(t, mdl.CreateSecurities(ctx, []*pb.Security{
		{Exchange: sp(exc), Symbol: sp(sym), Name: sp("Mediatek")},
	}))
	t.Cleanup(func() { rdb.Del(ctx, keyOfSecurityID(exc, sym)) })

	data, err := testdata.Mediatek()
	require.NoError(t, err)

	// Split: first batch covers 09:15–13:20, rest covers 09:00–09:14 and 13:21–13:25.
	start := timeOf("2026-02-26 09:15:00")
	end := timeOf("2026-02-26 13:20:00")
	var firstBatch, rest []*pb.OHLCV
	for _, r := range data {
		ts := r.GetTs().AsTime()
		if !ts.Before(start) && !ts.After(end) {
			firstBatch = append(firstBatch, r)
		} else {
			rest = append(rest, r)
		}
	}

	from := timeOf("2026-02-26")
	before := timeOf("2026-02-27")

	// Step 1: insert 09:15–13:20.
	require.NoError(t, mdl.CreateOHLCVs(ctx, exc, sym, Interval1m, firstBatch))

	// Step 2: first row = aggregated 09:15–09:29 (partial 09:00 bucket),
	//          last row = aggregated 13:00–13:20 (partial 13:00 bucket).
	rows, err := mdl.GetOHLCVs(ctx, exc, sym, Interval30m, from, before)
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	partial := aggregateOHLCVs(firstBatch, func(t time.Time) time.Time {
		return t.Truncate(30 * time.Minute)
	})
	// The timestamp of ohlcv_per_30min table is always aligned to 30-minute boundary,
	// so we need to adjust the expected timestamp of the first and last buckets.
	partial[0].Ts = timestamppb.New(start.Truncate(30 * time.Minute))
	assertBucket(t, partial[0], rows[0])
	assertBucket(t, partial[len(partial)-1], rows[len(rows)-1])

	// Step 3: insert the rest (09:00–09:14 and 13:21–13:25).
	require.NoError(t, mdl.CreateOHLCVs(ctx, exc, sym, Interval1m, rest))

	// Step 4: all rows in ohlcv_per_30min must match the full 30-minute
	//         aggregation of the entire dataset.
	all, err := mdl.GetOHLCVs(ctx, exc, sym, Interval30m, from, before)
	require.NoError(t, err)

	expected := aggregateOHLCVs(data, func(t time.Time) time.Time {
		return t.Truncate(30 * time.Minute)
	})
	require.Len(t, all, len(expected))
	for i, want := range expected {
		assertBucket(t, want, all[i])
	}
}

// TestOHLCVPerMinSizeEstimation inserts 5000 rows into ohlcv_per_min and
// reports the estimated table size. Run with -v to see the output.
func TestOHLCVPerMinSizeEstimation(t *testing.T) {
	t.Skip("manual run only")
	var (
		ctx  = context.Background()
		pool = connectPool(t)
		rdb  = connectRedis(t)
		mdl  = Connect(pool, rdb)
	)

	const (
		exch    = "twse_size_test"
		sym     = "SIZE01"
		numRows = 5000
	)

	// Clean up before and after.
	pool.Exec(ctx, "DELETE FROM exchanges WHERE abbr = $1", exch)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM exchanges WHERE abbr = $1", exch) })

	require.NoError(t, mdl.CreateExchange(ctx, &pb.Exchange{
		Abbr: sp(exch), Name: sp("Size Estimation Test Exchange"), Timezone: sp("UTC"),
	}))
	require.NoError(t, mdl.CreateSecurities(ctx, []*pb.Security{
		{Exchange: sp(exch), Symbol: sp(sym), Name: sp("Size Test Security")},
	}))

	// Generate 5000 synthetic 1-minute OHLCV rows starting from 2020-01-01.
	baseTime := time.Date(2020, 1, 1, 9, 0, 0, 0, time.UTC)
	rows := make([]*pb.OHLCV, numRows)
	price := 100.0
	for i := range rows {
		ts := baseTime.Add(time.Duration(i) * time.Minute)
		// Simple random-walk price simulation.
		if i%7 == 0 {
			price += 1.0
		} else if i%5 == 0 {
			price -= 0.5
		}
		o, h, l, c := price, price+0.5, price-0.5, price+0.1
		vol := uint64(100 + i%500)
		rows[i] = &pb.OHLCV{
			Ts:     timestamppb.New(ts),
			Open:   &o,
			High:   &h,
			Low:    &l,
			Close:  &c,
			Volume: &vol,
		}
	}

	// Insert in batches of 500 to avoid oversized transactions.
	const batchSize = 500
	for start := 0; start < numRows; start += batchSize {
		end := min(start+batchSize, numRows)
		require.NoError(t, mdl.CreateOHLCVs(ctx, exch, sym, Interval1m, rows[start:end]))
	}
	t.Logf("Inserted %d rows into ohlcv_per_min", numRows)

	// Query size of the parent table and its partitions.
	type tableSize struct {
		name  string
		total int64
		heap  int64
		index int64
		toast int64
	}

	// Query size stats for ohlcv_per_min (parent + all partitions).
	sizeQuery := `
		SELECT
			c.relname AS name,
			pg_total_relation_size(c.oid)  AS total,
			pg_relation_size(c.oid)        AS heap,
			pg_indexes_size(c.oid)         AS index,
			pg_total_relation_size(c.oid)
				- pg_relation_size(c.oid)
				- pg_indexes_size(c.oid)   AS toast
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public'
		  AND (c.relname = 'ohlcv_per_min'
		       OR c.relname LIKE 'ohlcv_per_min_%')
		  AND c.relkind IN ('r', 'p')
		ORDER BY c.relname`

	queryRows, err := pool.Query(ctx, sizeQuery)
	require.NoError(t, err)
	defer queryRows.Close()

	var sizes []tableSize
	for queryRows.Next() {
		var s tableSize
		require.NoError(t, queryRows.Scan(&s.name, &s.total, &s.heap, &s.index, &s.toast))
		sizes = append(sizes, s)
	}
	require.NoError(t, queryRows.Err())

	t.Logf("\n%-40s %12s %12s %12s %12s", "table", "total", "heap", "index", "toast")
	t.Logf("%s", fmt.Sprintf("%s", "----------------------------------------------------------------------------------------------------"))
	for _, s := range sizes {
		t.Logf("%-40s %12s %12s %12s %12s",
			s.name,
			formatBytes(s.total),
			formatBytes(s.heap),
			formatBytes(s.index),
			formatBytes(s.toast),
		)
	}

	// Per-row estimate based on ohlcv_per_min_default partition actual size.
	for _, s := range sizes {
		if s.name == "ohlcv_per_min_default" && s.total > 0 {
			perRow := float64(s.total) / float64(numRows)
			t.Logf("\nEstimated bytes per row (from default partition): %.1f bytes", perRow)
			t.Logf("Estimated size for 1M rows: %s", formatBytes(int64(perRow*1_000_000)))
			t.Logf("Estimated size for 10M rows: %s", formatBytes(int64(perRow*10_000_000)))
			t.Logf("Estimated size for 100M rows: %s", formatBytes(int64(perRow*100_000_000)))
		}
	}
}

// formatBytes formats a byte count into a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
