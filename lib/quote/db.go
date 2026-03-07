package quote

import (
	"context"
	"errors"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
)

// DB is a PostgreSQL-backed implementation of Model.
type DB struct {
	queries Querier
}

// Connect establishes a DB connection and returns a Model.
func Connect(db model.DBTX) Model {
	return &DB{queries: model.New(db)}
}

func (d *DB) CreateExchange(ctx context.Context, ex *pb.Exchange) error {
	if err := d.queries.InsertExchange(ctx, ex.GetAbbr(), ex.GetName(), ex.GetTimezone()); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (d *DB) GetExchanges(ctx context.Context) ([]*pb.Exchange, error) {
	rows, err := d.queries.GetExchanges(ctx)
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

func (d *DB) CreateSecurities(ctx context.Context, securities []*pb.Security) error {
	exchanges, err := d.queries.GetExchanges(ctx)
	if err != nil {
		return err
	}
	exchangeSet := make(map[string]bool, len(exchanges))
	for _, ex := range exchanges {
		exchangeSet[ex.Abbr] = true
	}

	params := make([]model.InsertSecuritiesParams, 0, len(securities))
	for _, sec := range securities {
		if !exchangeSet[sec.GetExchange()] {
			return ErrNotFound
		}
		params = append(params, model.InsertSecuritiesParams{
			Exchange: sec.GetExchange(),
			Symbol:   sec.GetSymbol(),
			Name:     sec.GetName(),
		})
	}

	if _, err := d.queries.InsertSecurities(ctx, params); err != nil {
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

func (d *DB) GetSecurities(ctx context.Context, exchange string) ([]*pb.Security, error) {
	rows, err := d.queries.GetSecurities(ctx, exchange)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		if _, err := d.queries.GetExchange(ctx, exchange); err != nil {
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

func (d *DB) CreateOHLCVs(ctx context.Context, exchange, symbol string, interval int64, ohlcvs []*pb.OHLCV) error {
	secs, err := d.queries.GetSecuritiesBySymbols(ctx, exchange, symbol)
	if err != nil || len(secs) == 0 {
		return ErrNotFound
	}
	secID := secs[0].ID

	if interval == Interval1d {
		return d.createOHLCVsPerDay(ctx, secID, ohlcvs)
	}
	return d.createOHLCVsPerMin(ctx, secID, ohlcvs)
}

func (d *DB) createOHLCVsPerDay(ctx context.Context, secID uuid.UUID, ohlcvs []*pb.OHLCV) error {
	params := make([]model.InsertOHLCVsPerDayParams, len(ohlcvs))
	for i, o := range ohlcvs {
		params[i] = model.InsertOHLCVsPerDayParams{
			SecID:  secID,
			Date:   pgtype.Date{Time: o.GetTs().AsTime(), Valid: true},
			Open:   floatToNumeric(o.GetOpen()),
			High:   floatToNumeric(o.GetHigh()),
			Low:    floatToNumeric(o.GetLow()),
			Close:  floatToNumeric(o.GetClose()),
			Volume: int64(o.GetVolume()),
		}
	}
	_, err := d.queries.InsertOHLCVsPerDay(ctx, params)
	return err
}

func (d *DB) createOHLCVsPerMin(ctx context.Context, secID uuid.UUID, ohlcvs []*pb.OHLCV) error {
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
	if _, err := d.queries.InsertOHLCVsPerMin(ctx, minParams); err != nil {
		return err
	}

	// Aggregate into 30-minute buckets and persist.
	thirtyMin := aggregateOHLCVs(ohlcvs, func(t time.Time) time.Time {
		return t.Truncate(30 * time.Minute)
	})
	thirtyMinParams := make([]model.InsertOHLCVsPer30MinParams, len(thirtyMin))
	for i, o := range thirtyMin {
		thirtyMinParams[i] = model.InsertOHLCVsPer30MinParams{
			SecID:  secID,
			Ts:     pgtype.Timestamp{Time: o.GetTs().AsTime(), Valid: true},
			Open:   floatToNumeric(o.GetOpen()),
			High:   floatToNumeric(o.GetHigh()),
			Low:    floatToNumeric(o.GetLow()),
			Close:  floatToNumeric(o.GetClose()),
			Volume: int64(o.GetVolume()),
		}
	}
	_, err := d.queries.InsertOHLCVsPer30Min(ctx, thirtyMinParams)
	return err
}

func (d *DB) GetOHLCVs(ctx context.Context, exchange, symbol string, interval int64, from, before time.Time) ([]*pb.OHLCV, error) {
	secs, err := d.queries.GetSecuritiesBySymbols(ctx, exchange, symbol)
	if err != nil || len(secs) == 0 {
		return nil, ErrNotFound
	}
	secID := secs[0].ID

	switch interval {
	case Interval1m, Interval5m:
		return d.getOHLCVsPerMin(ctx, secID, from, before, interval)
	case Interval30m, Interval1h:
		return d.getOHLCVsPer30Min(ctx, secID, from, before, interval)
	case Interval1d, Interval1w, Interval1M:
		return d.getOHLCVsPerDay(ctx, secID, from, before, interval)
	default:
		return nil, ErrInvalidArgument
	}
}

func (d *DB) getOHLCVsPerMin(ctx context.Context, secID uuid.UUID, from, before time.Time, interval int64) ([]*pb.OHLCV, error) {
	rows, err := d.queries.GetOHLCVsPerMin(ctx, secID,
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
	if interval == Interval5m {
		return aggregateOHLCVs(result, func(t time.Time) time.Time {
			return t.Truncate(5 * time.Minute)
		}), nil
	}
	return result, nil
}

func (d *DB) getOHLCVsPer30Min(ctx context.Context, secID uuid.UUID, from, before time.Time, interval int64) ([]*pb.OHLCV, error) {
	rows, err := d.queries.GetOHLCVsPer30Min(ctx, secID,
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

func (d *DB) getOHLCVsPerDay(ctx context.Context, secID uuid.UUID, from, before time.Time, interval int64) ([]*pb.OHLCV, error) {
	rows, err := d.queries.GetOHLCVsPerDay(ctx, secID,
		pgtype.Date{Time: from, Valid: true},
		pgtype.Date{Time: before, Valid: true},
	)
	if err != nil {
		return nil, err
	}
	result := make([]*pb.OHLCV, len(rows))
	for i, r := range rows {
		result[i] = ohlcvProto(r.Date.Time, r.Open, r.High, r.Low, r.Close, r.Volume)
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
