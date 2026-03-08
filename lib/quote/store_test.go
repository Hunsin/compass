package quote

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	mock "github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// testSecID is a fixed UUID used as a security ID in DB tests.
var testSecID = uuid.New()

func newTestDB(t *testing.T) (*store, *model.MockQuerier) {
	t.Helper()
	q := model.NewMockQuerier(t)
	return &store{queries: q}, q
}

// timeOf parse the timestamp in either "2006-01-02 15:04:05" or "2006-01-02" format.
func timeOf(ts string) time.Time {
	t, err := time.Parse(time.DateTime, ts)
	if err != nil {
		t, err = time.Parse(time.DateOnly, ts)
		if err != nil {
			panic("invalid time format: " + ts)
		}
	}
	return t
}

// ohlcvRow is a helper to create a pb.OHLCV from string and numeric literals.
func ohlcvRow(ts string, open, high, low, close_ float64, volume uint64) *pb.OHLCV {
	t := timeOf(ts)
	return &pb.OHLCV{
		Ts:     timestamppb.New(t),
		Open:   &open,
		High:   &high,
		Low:    &low,
		Close:  &close_,
		Volume: &volume,
	}
}

// data2317 holds daily OHLCV rows of Hon Hai.
var data2317 = []*pb.OHLCV{
	ohlcvRow("2025-10-01", 219.0, 222.0, 216.0, 219.0, 58_600_187),
	ohlcvRow("2025-10-02", 221.5, 226.0, 221.0, 225.5, 73_884_573),
	ohlcvRow("2025-10-03", 226.0, 228.0, 224.5, 226.5, 52_446_781),
	ohlcvRow("2025-10-07", 230.0, 231.5, 225.0, 228.5, 66_127_414),
	ohlcvRow("2025-10-08", 225.0, 225.5, 222.0, 225.0, 50_104_069),
	ohlcvRow("2025-10-09", 227.0, 227.0, 221.0, 221.5, 56_700_891),
	ohlcvRow("2025-10-13", 210.0, 216.0, 210.0, 213.0, 111_260_942),
	ohlcvRow("2025-10-14", 216.0, 217.5, 205.0, 206.0, 108_702_039),
	ohlcvRow("2025-10-15", 207.5, 208.0, 203.0, 206.5, 65_964_764),
	ohlcvRow("2025-10-16", 215.0, 225.5, 211.5, 223.0, 154_273_795),
	ohlcvRow("2025-10-17", 224.0, 230.0, 222.5, 226.5, 119_855_855),
	ohlcvRow("2025-10-20", 230.0, 239.5, 229.0, 238.5, 152_953_266),
	ohlcvRow("2025-10-21", 241.5, 245.5, 238.0, 239.0, 120_698_819),
	ohlcvRow("2025-10-22", 238.0, 243.0, 236.5, 242.5, 71_609_786),
	ohlcvRow("2025-10-23", 240.0, 240.5, 238.0, 239.0, 50_955_830),
	ohlcvRow("2025-10-27", 249.0, 249.0, 243.0, 246.5, 80_708_542),
	ohlcvRow("2025-10-28", 248.0, 250.0, 246.5, 250.0, 73_662_546),
	ohlcvRow("2025-10-29", 254.0, 263.0, 253.0, 260.0, 125_297_120),
	ohlcvRow("2025-10-30", 262.5, 265.0, 256.5, 262.0, 108_445_874),
	ohlcvRow("2025-10-31", 261.0, 262.0, 257.0, 257.5, 63_119_828),
	ohlcvRow("2025-11-03", 255.0, 256.5, 250.0, 251.5, 97_190_219),
	ohlcvRow("2025-11-04", 252.0, 255.0, 244.0, 244.0, 95_832_447),
	ohlcvRow("2025-11-05", 239.0, 248.5, 237.0, 247.0, 114_664_242),
	ohlcvRow("2025-11-06", 249.5, 250.5, 246.0, 248.0, 65_794_425),
	ohlcvRow("2025-11-07", 246.0, 250.5, 244.0, 244.0, 50_319_199),
	ohlcvRow("2025-11-10", 245.0, 250.0, 244.0, 249.5, 51_628_691),
	ohlcvRow("2025-11-11", 251.0, 252.0, 246.5, 246.5, 50_821_377),
	ohlcvRow("2025-11-12", 248.0, 252.0, 245.0, 251.0, 65_085_447),
	ohlcvRow("2025-11-13", 256.0, 256.0, 247.0, 252.0, 90_541_908),
	ohlcvRow("2025-11-14", 246.5, 247.0, 241.0, 241.0, 112_889_692),
	ohlcvRow("2025-11-17", 241.0, 242.0, 235.0, 236.0, 84_874_596),
	ohlcvRow("2025-11-18", 231.5, 237.0, 230.0, 230.0, 77_859_170),
	ohlcvRow("2025-11-19", 226.5, 230.0, 226.5, 229.0, 70_909_837),
	ohlcvRow("2025-11-20", 237.0, 238.0, 232.5, 236.5, 78_464_576),
	ohlcvRow("2025-11-21", 226.0, 229.5, 225.0, 225.0, 97_471_576),
	ohlcvRow("2025-11-24", 226.5, 227.0, 220.0, 220.0, 100_270_276),
	ohlcvRow("2025-11-25", 225.0, 226.0, 219.0, 219.0, 73_643_136),
	ohlcvRow("2025-11-26", 221.0, 228.0, 221.0, 227.0, 68_461_990),
	ohlcvRow("2025-11-27", 229.0, 232.0, 226.0, 231.0, 63_038_658),
	ohlcvRow("2025-11-28", 229.5, 229.5, 225.0, 225.5, 56_939_840),
}

func TestWeekBucket(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "Monday stays Monday",
			in:   timeOf("2025-12-08 10:30:00"),
			want: timeOf("2025-12-08 00:00:00"),
		},
		{
			name: "Wednesday maps to Monday",
			in:   timeOf("2025-12-10 09:00:00"),
			want: timeOf("2025-12-08 00:00:00"),
		},
		{
			name: "Sunday maps to Monday",
			in:   timeOf("2025-12-14 15:00:00"),
			want: timeOf("2025-12-08 00:00:00"),
		},
		{
			name: "Friday maps to Monday",
			in:   timeOf("2025-12-12 12:00:00"),
			want: timeOf("2025-12-08 00:00:00"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := weekBucket(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("weekBucket(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestMonthBucket(t *testing.T) {
	tests := []struct {
		name string
		in   time.Time
		want time.Time
	}{
		{
			name: "first of month unchanged",
			in:   timeOf("2025-12-01 00:00:00"),
			want: timeOf("2025-12-01 00:00:00"),
		},
		{
			name: "mid-month maps to first",
			in:   timeOf("2025-12-15 09:30:00"),
			want: timeOf("2025-12-01 00:00:00"),
		},
		{
			name: "last day maps to first",
			in:   timeOf("2025-11-30 23:59:00"),
			want: timeOf("2025-11-01 00:00:00"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := monthBucket(tc.in)
			if !got.Equal(tc.want) {
				t.Errorf("monthBucket(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestAggregateWeeklyOHLCVs(t *testing.T) {
	got := aggregateOHLCVs(data2317, weekBucket)

	want := []struct {
		date                    string
		open, high, low, close_ float64
		volume                  uint64
	}{
		// Oct 01–03
		{"2025-10-01", 219.0, 228.0, 216.0, 226.5, 184_931_541},
		// Oct 07–09; Monday and Friday are holidays
		{"2025-10-07", 230.0, 231.5, 221.0, 221.5, 172_932_374},
		// Oct 13–17
		{"2025-10-13", 210.0, 230.0, 203.0, 226.5, 560_057_395},
		// Oct 20–23; Friday is Taiwan's Retrocession Day
		{"2025-10-20", 230.0, 245.5, 229.0, 239.0, 396_217_701},
		// Oct 27-31
		{"2025-10-27", 249.0, 265.0, 243.0, 257.5, 451_233_910},
		// Nov 03-07
		{"2025-11-03", 255.0, 256.5, 237.0, 244.0, 423_800_532},
		// Nov 10-14
		{"2025-11-10", 245.0, 256.0, 241.0, 241.0, 370_967_115},
		// Nov 17-21
		{"2025-11-17", 241.0, 242.0, 225.0, 225.0, 409_579_755},
		// Nov 24-28
		{"2025-11-24", 226.5, 232.0, 219.0, 225.5, 362_353_900},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d weekly buckets, want %d", len(got), len(want))
	}

	for i, w := range want {
		g := got[i]
		wt := timeOf(w.date)
		if !g.GetTs().AsTime().Equal(wt) {
			t.Errorf("week[%d] ts: got %v, want %v", i, g.GetTs().AsTime(), wt)
		}
		if g.GetOpen() != w.open {
			t.Errorf("week[%d] open: got %v, want %v", i, g.GetOpen(), w.open)
		}
		if g.GetHigh() != w.high {
			t.Errorf("week[%d] high: got %v, want %v", i, g.GetHigh(), w.high)
		}
		if g.GetLow() != w.low {
			t.Errorf("week[%d] low: got %v, want %v", i, g.GetLow(), w.low)
		}
		if g.GetClose() != w.close_ {
			t.Errorf("week[%d] close: got %v, want %v", i, g.GetClose(), w.close_)
		}
		if g.GetVolume() != w.volume {
			t.Errorf("week[%d] volume: got %v, want %v", i, g.GetVolume(), w.volume)
		}
	}
}

func TestAggregateMonthlyOHLCVs(t *testing.T) {
	got := aggregateOHLCVs(data2317, monthBucket)

	want := []struct {
		date                    string
		open, high, low, close_ float64
		volume                  uint64
	}{
		// October 2025
		{"2025-10-01", 219.0, 265.0, 203.0, 257.5, 1_765_372_921},
		// November 2025; 11-03 is the first trading date of the month
		{"2025-11-03", 255.0, 256.5, 219.0, 225.5, 1_566_701_302},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d monthly buckets, want %d", len(got), len(want))
	}

	for i, w := range want {
		g := got[i]
		wt := timeOf(w.date)
		if !g.GetTs().AsTime().Equal(wt) {
			t.Errorf("month[%d] ts: got %v, want %v", i, g.GetTs().AsTime(), wt)
		}
		if g.GetOpen() != w.open {
			t.Errorf("month[%d] open: got %v, want %v", i, g.GetOpen(), w.open)
		}
		if g.GetHigh() != w.high {
			t.Errorf("month[%d] high: got %v, want %v", i, g.GetHigh(), w.high)
		}
		if g.GetLow() != w.low {
			t.Errorf("month[%d] low: got %v, want %v", i, g.GetLow(), w.low)
		}
		if g.GetClose() != w.close_ {
			t.Errorf("month[%d] close: got %v, want %v", i, g.GetClose(), w.close_)
		}
		if g.GetVolume() != w.volume {
			t.Errorf("month[%d] volume: got %v, want %v", i, g.GetVolume(), w.volume)
		}
	}
}

func TestAggregateOHLCVs_Empty(t *testing.T) {
	got := aggregateOHLCVs(nil, func(t time.Time) time.Time { return t })
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d rows", len(got))
	}
}

// --- DB unit tests ---

func TestDB_CreateExchange(t *testing.T) {
	ctx := context.Background()
	abbr, name, tz := "TSE", "Taiwan Stock Exchange", "Asia/Taipei"
	ex := &pb.Exchange{Abbr: &abbr, Name: &name, Timezone: &tz}

	t.Run("success", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().InsertExchange(ctx, abbr, name, tz).Return(nil)
		if err := db.CreateExchange(ctx, ex); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().InsertExchange(ctx, abbr, name, tz).Return(&pgconn.PgError{Code: "23505"})
		if !errors.Is(db.CreateExchange(ctx, ex), ErrAlreadyExists) {
			t.Error("want ErrAlreadyExists")
		}
	})

	t.Run("other error propagated", func(t *testing.T) {
		db, q := newTestDB(t)
		dbErr := errors.New("connection reset")
		q.EXPECT().InsertExchange(ctx, abbr, name, tz).Return(dbErr)
		if !errors.Is(db.CreateExchange(ctx, ex), dbErr) {
			t.Error("want original error")
		}
	})
}

func TestDB_GetExchanges(t *testing.T) {
	ctx := context.Background()

	t.Run("returns exchanges", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetExchanges(ctx).Return([]model.Exchange{
			{Abbr: "TSE", Name: "Taiwan SE", Timezone: "Asia/Taipei"},
		}, nil)
		got, err := db.GetExchanges(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].GetAbbr() != "TSE" {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("error propagated", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetExchanges(ctx).Return(nil, errors.New("db error"))
		got, err := db.GetExchanges(ctx)
		if err == nil || got != nil {
			t.Error("want error and nil result")
		}
	})
}

func TestDB_CreateSecurities(t *testing.T) {
	ctx := context.Background()
	exch, sym, secName := "TSE", "2317", "Hon Hai"
	sec := &pb.Security{Exchange: &exch, Symbol: &sym, Name: &secName}
	params := []model.InsertSecuritiesParams{{Exchange: exch, Symbol: sym, Name: secName}}

	t.Run("exchange not found", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetExchanges(ctx).Return([]model.Exchange{}, nil)
		if !errors.Is(db.CreateSecurities(ctx, []*pb.Security{sec}), ErrNotFound) {
			t.Error("want ErrNotFound")
		}
	})

	t.Run("success", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetExchanges(ctx).Return([]model.Exchange{{Abbr: exch}}, nil)
		q.EXPECT().InsertSecurities(ctx, params).Return(int64(1), nil)
		if err := db.CreateSecurities(ctx, []*pb.Security{sec}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("already exists", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetExchanges(ctx).Return([]model.Exchange{{Abbr: exch}}, nil)
		q.EXPECT().InsertSecurities(ctx, mock.Anything).Return(int64(0), &pgconn.PgError{Code: "23505"})
		if !errors.Is(db.CreateSecurities(ctx, []*pb.Security{sec}), ErrAlreadyExists) {
			t.Error("want ErrAlreadyExists")
		}
	})

	t.Run("foreign key violation", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetExchanges(ctx).Return([]model.Exchange{{Abbr: exch}}, nil)
		q.EXPECT().InsertSecurities(ctx, mock.Anything).Return(int64(0), &pgconn.PgError{Code: "23503"})
		if !errors.Is(db.CreateSecurities(ctx, []*pb.Security{sec}), ErrNotFound) {
			t.Error("want ErrNotFound")
		}
	})
}

func TestDB_GetSecurities(t *testing.T) {
	ctx := context.Background()
	const exch = "TSE"

	t.Run("returns securities", func(t *testing.T) {
		db, q := newTestDB(t)
		sym, name := "2317", "Hon Hai"
		q.EXPECT().GetSecurities(ctx, exch).Return([]model.Security{
			{ID: testSecID, Exchange: exch, Symbol: sym, Name: name},
		}, nil)
		got, err := db.GetSecurities(ctx, exch)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].GetSymbol() != sym {
			t.Errorf("unexpected result: %v", got)
		}
	})

	t.Run("empty result when exchange exists", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecurities(ctx, exch).Return([]model.Security{}, nil)
		q.EXPECT().GetExchange(ctx, exch).Return(model.Exchange{Abbr: exch}, nil)
		got, err := db.GetSecurities(ctx, exch)
		if err != nil || got != nil {
			t.Errorf("got %v, %v; want nil, nil", got, err)
		}
	})

	t.Run("exchange not found", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecurities(ctx, exch).Return([]model.Security{}, nil)
		q.EXPECT().GetExchange(ctx, exch).Return(model.Exchange{}, errors.New("no rows"))
		_, err := db.GetSecurities(ctx, exch)
		if !errors.Is(err, ErrNotFound) {
			t.Error("want ErrNotFound")
		}
	})
}

func TestDB_CreateOHLCVs(t *testing.T) {
	ctx := context.Background()
	const exch, sym = "TSE", "2317"
	secs := []model.Security{{ID: testSecID}}
	row := ohlcvRow("2026-01-02 00:00:00", 232.0, 233.5, 229.0, 232.0, 58_776_015)

	t.Run("security not found", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(nil, nil)
		if !errors.Is(db.CreateOHLCVs(ctx, exch, sym, Interval1d, []*pb.OHLCV{row}), ErrNotFound) {
			t.Error("want ErrNotFound")
		}
	})

	t.Run("day interval", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(secs, nil)
		q.EXPECT().InsertOHLCVsPerDay(ctx, mock.Anything).Return(int64(1), nil)
		if err := db.CreateOHLCVs(ctx, exch, sym, Interval1d, []*pb.OHLCV{row}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("minute interval inserts per-min and 30-min", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(secs, nil)
		q.EXPECT().InsertOHLCVsPerMin(ctx, mock.Anything).Return(int64(1), nil)
		q.EXPECT().InsertOHLCVsPer30Min(ctx, mock.Anything).Return(int64(1), nil)
		if err := db.CreateOHLCVs(ctx, exch, sym, Interval1m, []*pb.OHLCV{row}); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestDB_GetOHLCVs(t *testing.T) {
	ctx := context.Background()
	const exch, sym = "TSE", "2317"
	from := timeOf("2026-01-01 00:00:00")
	before := timeOf("2026-02-01 00:00:00")
	secs := []model.Security{{ID: testSecID}}

	t.Run("security not found", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(nil, nil)
		_, err := db.GetOHLCVs(ctx, exch, sym, Interval1d, from, before)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("day interval returns rows as-is", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(secs, nil)
		q.EXPECT().GetOHLCVsPerDay(ctx, testSecID,
			civil.DateOf(from),
			civil.DateOf(before),
		).Return([]model.OHLCVperDay{
			{
				Date:   civil.Date{Year: 2026, Month: time.January, Day: 2},
				Open:   floatToNumeric(232.0),
				High:   floatToNumeric(233.5),
				Low:    floatToNumeric(229.0),
				Close:  floatToNumeric(232.0),
				Volume: 58_776_015,
			},
		}, nil)
		got, err := db.GetOHLCVs(ctx, exch, sym, Interval1d, from, before)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].GetOpen() != 232.0 {
			t.Errorf("unexpected result")
		}
	})

	t.Run("week interval aggregates daily rows", func(t *testing.T) {
		db, q := newTestDB(t)
		// Jan 5 and Jan 6 fall in the same ISO week (starting Jan 5).
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(secs, nil)
		q.EXPECT().GetOHLCVsPerDay(ctx, testSecID,
			civil.DateOf(from),
			civil.DateOf(before),
		).Return([]model.OHLCVperDay{
			{Date: civil.DateOf(timeOf("2026-01-05")), Open: floatToNumeric(234.5), High: floatToNumeric(236.0), Low: floatToNumeric(233.5), Close: floatToNumeric(234.5), Volume: 64_697_110},
			{Date: civil.DateOf(timeOf("2026-01-06")), Open: floatToNumeric(237.0), High: floatToNumeric(239.0), Low: floatToNumeric(232.5), Close: floatToNumeric(236.0), Volume: 68_919_645},
		}, nil)
		got, err := db.GetOHLCVs(ctx, exch, sym, Interval1w, from, before)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("got %d buckets, want 1", len(got))
		}
		if got[0].GetHigh() != 239.0 {
			t.Errorf("high: got %v, want 239.0", got[0].GetHigh())
		}
		if got[0].GetVolume() != 64_697_110+68_919_645 {
			t.Errorf("volume: got %v, want %v", got[0].GetVolume(), uint64(64_697_110+68_919_645))
		}
	})

	t.Run("minute interval returns rows as-is", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(secs, nil)
		q.EXPECT().GetOHLCVsPerMin(ctx, testSecID,
			pgtype.Timestamp{Time: from, Valid: true},
			pgtype.Timestamp{Time: before, Valid: true},
		).Return([]model.OHLCVperMin{
			{
				Ts:     pgtype.Timestamp{Time: timeOf("2026-01-02 09:01:00"), Valid: true},
				Open:   floatToNumeric(232.0),
				High:   floatToNumeric(233.0),
				Low:    floatToNumeric(231.5),
				Close:  floatToNumeric(232.5),
				Volume: 1_000_000,
			},
		}, nil)
		got, err := db.GetOHLCVs(ctx, exch, sym, Interval1m, from, before)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].GetOpen() != 232.0 {
			t.Errorf("unexpected result")
		}
	})

	t.Run("30-minute interval returns rows as-is", func(t *testing.T) {
		db, q := newTestDB(t)
		q.EXPECT().GetSecuritiesBySymbols(ctx, exch, sym).Return(secs, nil)
		q.EXPECT().GetOHLCVsPer30Min(ctx, testSecID,
			pgtype.Timestamp{Time: from, Valid: true},
			pgtype.Timestamp{Time: before, Valid: true},
		).Return([]model.OHLCVper30Min{
			{
				Ts:     pgtype.Timestamp{Time: timeOf("2026-01-02 09:00:00"), Valid: true},
				Open:   floatToNumeric(232.0),
				High:   floatToNumeric(233.5),
				Low:    floatToNumeric(231.0),
				Close:  floatToNumeric(233.0),
				Volume: 5_000_000,
			},
		}, nil)
		got, err := db.GetOHLCVs(ctx, exch, sym, Interval30m, from, before)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 1 || got[0].GetOpen() != 232.0 {
			t.Errorf("unexpected result")
		}
	})
}
