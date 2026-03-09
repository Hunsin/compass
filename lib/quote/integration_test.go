//go:build integration

package quote

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	pool := connectPool(t)
	ctx := context.Background()
	s := Connect(pool).(*store)

	const exch, sym = "twse", "2454"

	// Ensure clean state before and after.
	pool.Exec(ctx, "DELETE FROM exchanges WHERE abbr = $1", exch)
	t.Cleanup(func() { pool.Exec(ctx, "DELETE FROM exchanges WHERE abbr = $1", exch) })

	require.NoError(t, s.CreateExchange(ctx, &pb.Exchange{
		Abbr: sp(exch), Name: sp("Integration Test"), Timezone: sp("UTC"),
	}))
	require.NoError(t, s.CreateSecurities(ctx, []*pb.Security{
		{Exchange: sp(exch), Symbol: sp(sym), Name: sp("Mediatek")},
	}))

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
	require.NoError(t, s.CreateOHLCVs(ctx, exch, sym, Interval1m, firstBatch))

	// Step 2: first row = aggregated 09:15–09:29 (partial 09:00 bucket),
	//          last row = aggregated 13:00–13:20 (partial 13:00 bucket).
	rows, err := s.GetOHLCVs(ctx, exch, sym, Interval30m, from, before)
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
	require.NoError(t, s.CreateOHLCVs(ctx, exch, sym, Interval1m, rest))

	// Step 4: all rows in ohlcv_per_30min must match the full 30-minute
	//         aggregation of the entire dataset.
	all, err := s.GetOHLCVs(ctx, exch, sym, Interval30m, from, before)
	require.NoError(t, err)

	expected := aggregateOHLCVs(data, func(t time.Time) time.Time {
		return t.Truncate(30 * time.Minute)
	})
	require.Len(t, all, len(expected))
	for i, want := range expected {
		assertBucket(t, want, all[i])
	}
}
