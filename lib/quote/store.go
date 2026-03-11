package quote

import (
	"context"
	"errors"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

type DBTX interface {
	model.DBTX
	Begin(context.Context) (pgx.Tx, error)
}

// store is a PostgreSQL-backed implementation of Model.
type store struct {
	db      DBTX
	queries model.Querier
}

// Connect establishes a DB connection and returns a Model.
func Connect(db DBTX) Model {
	return &store{db: db, queries: model.New(db)}
}

func (s *store) CreateExchange(ctx context.Context, ex *pb.Exchange) error {
	abbr := strings.ToLower(ex.GetAbbr())
	if err := s.queries.InsertExchange(ctx, abbr, ex.GetName(), ex.GetTimezone()); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (s *store) GetExchanges(ctx context.Context) ([]*pb.Exchange, error) {
	rows, err := s.queries.GetExchanges(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.Exchange, len(rows))
	for i, r := range rows {
		abbr := r.Abbr
		name := r.Name
		tz := r.Timezone
		result[i] = &pb.Exchange{Abbr: &abbr, Name: &name, Timezone: &tz}
	}
	return result, nil
}

func (s *store) CreateSecurities(ctx context.Context, securities []*pb.Security) error {
	exchanges, err := s.queries.GetExchanges(ctx)
	if err != nil {
		return err
	}
	m := make(map[string]bool, len(exchanges))
	for _, ex := range exchanges {
		m[strings.ToLower(ex.Abbr)] = true
	}

	params := make([]model.InsertSecuritiesParams, 0, len(securities))
	for _, sec := range securities {
		abbr := strings.ToLower(sec.GetExchange())
		if !m[abbr] {
			return ErrNotFound
		}
		params = append(params, model.InsertSecuritiesParams{
			Exchange: abbr,
			Symbol:   strings.ToUpper(sec.GetSymbol()),
			Name:     sec.GetName(),
		})
	}

	if _, err := s.queries.InsertSecurities(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return ErrAlreadyExists
			case "23503":
				return ErrNotFound
			}
		}
		return err
	}
	return nil
}

func (s *store) GetSecurities(ctx context.Context, exchange string) ([]*pb.Security, error) {
	exchange = strings.ToLower(exchange)

	rows, err := s.queries.GetSecurities(ctx, exchange)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		if _, err := s.queries.GetExchange(ctx, exchange); err != nil {
			return nil, ErrNotFound
		}
		return nil, nil
	}

	result := make([]*pb.Security, len(rows))
	for i, r := range rows {
		exch := r.Exchange
		sym := r.Symbol
		name := r.Name
		result[i] = &pb.Security{Exchange: &exch, Symbol: &sym, Name: &name}
	}
	return result, nil
}

func (s *store) CreateOHLCVs(ctx context.Context, exchange, symbol string, interval int64, ohlcvs []*pb.OHLCV) error {
	exchange = strings.ToLower(exchange)
	symbol = strings.ToUpper(symbol)

	secs, err := s.queries.GetSecuritiesBySymbols(ctx, exchange, []string{symbol})
	if err != nil {
		return err
	}
	if len(secs) == 0 {
		return ErrNotFound
	}
	secID := secs[0].ID

	switch interval {
	case Interval1d:
		return s.createOHLCVsPerDay(ctx, secID, ohlcvs)
	case Interval1m:
		return s.createOHLCVsPerMin(ctx, secID, ohlcvs)
	default:
		return ErrInvalidArgument
	}
}

func (s *store) createOHLCVsPerDay(ctx context.Context, secID uuid.UUID, ohlcvs []*pb.OHLCV) error {
	params := make([]model.InsertOHLCVsPerDayParams, len(ohlcvs))
	for i, o := range ohlcvs {
		params[i] = model.InsertOHLCVsPerDayParams{
			SecID:  secID,
			Date:   civil.DateOf(o.GetTs().AsTime()),
			Open:   floatToNumeric(o.GetOpen()),
			High:   floatToNumeric(o.GetHigh()),
			Low:    floatToNumeric(o.GetLow()),
			Close:  floatToNumeric(o.GetClose()),
			Volume: int64(o.GetVolume()),
		}
	}
	_, err := s.queries.InsertOHLCVsPerDay(ctx, params)
	return err
}

func (s *store) createOHLCVsPerMin(ctx context.Context, secID uuid.UUID, ohlcvs []*pb.OHLCV) error {
	minParams := make([]model.InsertOHLCVsPerMinParams, len(ohlcvs))
	for i, o := range ohlcvs {
		t := o.GetTs().AsTime().Truncate(time.Minute)
		minParams[i] = model.InsertOHLCVsPerMinParams{
			SecID:  secID,
			Ts:     pgtype.Timestamp{Time: t, Valid: true},
			Open:   floatToNumeric(o.GetOpen()),
			High:   floatToNumeric(o.GetHigh()),
			Low:    floatToNumeric(o.GetLow()),
			Close:  floatToNumeric(o.GetClose()),
			Volume: int64(o.GetVolume()),
		}
	}
	if _, err := s.queries.InsertOHLCVsPerMin(ctx, minParams); err != nil {
		return err
	}

	// Aggregate into 30-minute buckets and upsert.
	type bucket30 struct {
		open   float64
		high   float64
		low    float64
		close  float64
		volume uint64
		minTs  time.Time
		maxTs  time.Time
	}
	buckets := make(map[time.Time]*bucket30)
	var order []time.Time
	for _, o := range ohlcvs {
		t := o.GetTs().AsTime().Truncate(time.Minute)
		k := t.Truncate(30 * time.Minute)
		if b, ok := buckets[k]; !ok {
			buckets[k] = &bucket30{
				open: o.GetOpen(), high: o.GetHigh(),
				low: o.GetLow(), close: o.GetClose(),
				volume: o.GetVolume(), minTs: t, maxTs: t,
			}
			order = append(order, k)
		} else {
			if o.GetHigh() > b.high {
				b.high = o.GetHigh()
			}
			if o.GetLow() < b.low {
				b.low = o.GetLow()
			}
			b.volume += o.GetVolume()
			if t.Before(b.minTs) {
				b.minTs = t
				b.open = o.GetOpen()
			}
			if t.After(b.maxTs) {
				b.maxTs = t
				b.close = o.GetClose()
			}
		}
	}
	params := make([]model.UpsertOHLCVPer30MinParams, 0, len(order))
	for _, k := range order {
		b := buckets[k]

		// TODO: currently handle special case at 13:25 for Taiwanese market.
		// Need to find a more flexible way to determine if the last bucket is partial.
		isLast := b.maxTs.Equal(k.Add(29*time.Minute)) ||
			b.maxTs.Hour() == 13 && b.maxTs.Minute() == 25

		params = append(params, model.UpsertOHLCVPer30MinParams{
			SecID:   secID,
			Ts:      pgtype.Timestamp{Time: k, Valid: true},
			Open:    floatToNumeric(b.open),
			High:    floatToNumeric(b.high),
			Low:     floatToNumeric(b.low),
			Close:   floatToNumeric(b.close),
			Volume:  int64(b.volume),
			IsFirst: b.minTs.Equal(k),
			IsLast:  isLast,
		})
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tq := model.New(tx)
	for _, p := range params {
		if err := tq.UpsertOHLCVPer30Min(ctx, p); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *store) GetOHLCVs(ctx context.Context, exchange, symbol string, interval int64, from, before time.Time) ([]*pb.OHLCV, error) {
	exchange = strings.ToLower(exchange)
	symbol = strings.ToUpper(symbol)
	secs, err := s.queries.GetSecuritiesBySymbols(ctx, exchange, []string{symbol})
	if err != nil {
		return nil, err
	}
	if len(secs) == 0 {
		return nil, ErrNotFound
	}
	secID := secs[0].ID

	switch interval {
	case Interval1m, Interval5m:
		return s.getOHLCVsPerMin(ctx, secID, from, before, interval)
	case Interval30m, Interval1h:
		return s.getOHLCVsPer30Min(ctx, secID, from, before, interval)
	case Interval1d, Interval1w, Interval1M:
		return s.getOHLCVsPerDay(ctx, secID, from, before, interval)
	default:
		return nil, ErrInvalidArgument
	}
}

func (s *store) getOHLCVsPerMin(ctx context.Context, secID uuid.UUID, from, before time.Time, interval int64) ([]*pb.OHLCV, error) {
	rows, err := s.queries.GetOHLCVsPerMin(ctx, secID,
		pgtype.Timestamp{Time: from, Valid: true},
		pgtype.Timestamp{Time: before, Valid: true},
	)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.OHLCV, len(rows))
	for i, r := range rows {
		result[i] = ohlcvProto(r.Ts.Time, r.Open, r.High, r.Low, r.Close, r.Volume)
	}

	if interval == Interval1m {
		return result, nil
	}

	var d time.Duration
	switch interval {
	case Interval5m:
		d = 5 * time.Minute
	case Interval30m:
		d = 30 * time.Minute
	case Interval1h:
		d = time.Hour
	default:
		return nil, ErrInvalidArgument
	}

	return aggregateOHLCVs(result, func(t time.Time) time.Time {
		return t.Truncate(d)
	}), nil
}

func (s *store) getOHLCVsPer30Min(ctx context.Context, secID uuid.UUID, from, before time.Time, interval int64) ([]*pb.OHLCV, error) {
	// If from/before is not aligned to 30-minute boundary, we need to aggregate from per-minute data.
	if from.After(from.Truncate(30*time.Minute)) ||
		before.After(before.Truncate(30*time.Minute)) {
		return s.getOHLCVsPerMin(ctx, secID, from, before, interval)
	}

	rows, err := s.queries.GetOHLCVsPer30Min(ctx, secID,
		pgtype.Timestamp{Time: from, Valid: true},
		pgtype.Timestamp{Time: before, Valid: true},
	)
	if err != nil {
		return nil, err
	}

	result := make([]*pb.OHLCV, len(rows))
	for i, r := range rows {
		result[i] = ohlcvProto(r.Ts.Time, r.Open, r.High, r.Low, r.Close, r.Volume)
	}
	if interval == Interval1h {
		return aggregateOHLCVs(result, func(t time.Time) time.Time {
			return t.Truncate(time.Hour)
		}), nil
	}
	return result, nil
}

func (s *store) getOHLCVsPerDay(ctx context.Context, secID uuid.UUID, from, before time.Time, interval int64) ([]*pb.OHLCV, error) {
	rows, err := s.queries.GetOHLCVsPerDay(ctx, secID,
		civil.DateOf(from),
		civil.DateOf(before),
	)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.OHLCV, len(rows))
	for i, r := range rows {
		result[i] = ohlcvProto(r.Date.In(time.UTC), r.Open, r.High, r.Low, r.Close, r.Volume)
	}
	if interval == Interval1d {
		return result, nil
	}
	bucketFn := weekBucket
	if interval == Interval1M {
		bucketFn = monthBucket
	}
	return aggregateOHLCVs(result, bucketFn), nil
}

// aggregateOHLCVs groups OHLCV rows by bucket key and aggregates them.
// Rows must be sorted by time ascending.
func aggregateOHLCVs(rows []*pb.OHLCV, bucket func(time.Time) time.Time) []*pb.OHLCV {
	type agg struct {
		ts     time.Time
		open   float64
		high   float64
		low    float64
		close  float64
		volume uint64
	}

	buckets := make(map[time.Time]*agg)
	var order []time.Time

	for _, row := range rows {
		k := bucket(row.GetTs().AsTime())
		if b, ok := buckets[k]; !ok {
			buckets[k] = &agg{
				ts:     row.GetTs().AsTime(),
				open:   row.GetOpen(),
				high:   row.GetHigh(),
				low:    row.GetLow(),
				close:  row.GetClose(),
				volume: row.GetVolume(),
			}
			order = append(order, k)
		} else {
			if row.GetHigh() > b.high {
				b.high = row.GetHigh()
			}
			if row.GetLow() < b.low {
				b.low = row.GetLow()
			}
			b.close = row.GetClose()
			b.volume += row.GetVolume()
		}
	}

	result := make([]*pb.OHLCV, 0, len(order))
	for _, k := range order {
		b := buckets[k]
		o := b.open
		h := b.high
		l := b.low
		c := b.close
		v := b.volume
		result = append(result, &pb.OHLCV{
			Ts:     timestamppb.New(b.ts),
			Open:   &o,
			High:   &h,
			Low:    &l,
			Close:  &c,
			Volume: &v,
		})
	}
	return result
}

// weekBucket returns the Monday of the ISO week containing t.
func weekBucket(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// monthBucket returns the first day of the month containing t.
func monthBucket(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// ohlcvProto constructs a *pb.OHLCV from DB row fields.
func ohlcvProto(ts time.Time, open, high, low, close_ pgtype.Numeric, volume int64) *pb.OHLCV {
	o := numericToFloat(open)
	h := numericToFloat(high)
	l := numericToFloat(low)
	c := numericToFloat(close_)
	v := uint64(volume)
	return &pb.OHLCV{
		Ts:     timestamppb.New(ts),
		Open:   &o,
		High:   &h,
		Low:    &l,
		Close:  &c,
		Volume: &v,
	}
}

func floatToNumeric(f float64) pgtype.Numeric {
	var n pgtype.Numeric

	s := strconv.FormatFloat(f, 'f', -1, 64)
	if err := n.Scan(s); err != nil {
		panic("failed to convert float to numeric: " + err.Error())
	}
	return n
}

func numericToFloat(n pgtype.Numeric) float64 {
	if !n.Valid || n.NaN || n.Int == nil {
		return 0
	}
	f, _ := new(big.Float).SetInt(n.Int).Float64()
	if n.Exp != 0 {
		f *= math.Pow10(int(n.Exp))
	}
	return f
}
