package quote

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
)

const (
	interval1m  = 60
	interval5m  = 300
	interval30m = 1800
	interval1h  = 3600
	interval1d  = 86400
	interval1w  = 604800
	interval1M  = 2592000
)

func (s *Service) CreateOHLCVAs(ctx context.Context, req *pb.CreateOHLCVAsRequest) (*emptypb.Empty, error) {
	if req.GetExchange() == "" || req.GetSymbol() == "" {
		return nil, status.Error(codes.InvalidArgument, "exchange and symbol are required")
	}
	if req.Interval == nil {
		return nil, status.Error(codes.InvalidArgument, "interval is required")
	}
	if len(req.Ohlcva) == 0 {
		return nil, status.Error(codes.InvalidArgument, "ohlcva data is required")
	}

	intervalSecs := req.Interval.Seconds
	if intervalSecs != interval1m && intervalSecs != interval1d {
		return nil, status.Error(codes.InvalidArgument, "interval must be 1m or 1d")
	}

	secs, err := s.db.GetSecuritiesBySymbols(ctx, req.GetExchange(), req.GetSymbol())
	if err != nil || len(secs) == 0 {
		return nil, status.Error(codes.NotFound, "security not found")
	}
	secID := secs[0].ID

	if intervalSecs == interval1d {
		return s.createOHLCVAsPerDay(ctx, secID, req.Ohlcva)
	}
	return s.createOHLCVAsPerMin(ctx, secID, req.Ohlcva)
}

func (s *Service) createOHLCVAsPerDay(ctx context.Context, secID pgtype.UUID, data []*pb.OHLCVA) (*emptypb.Empty, error) {
	params := make([]model.InsertOHLCVAsPerDayParams, 0, len(data))
	for _, o := range data {
		if o.GetTs() == nil {
			return nil, status.Error(codes.InvalidArgument, "timestamp is required")
		}
		t := o.GetTs().AsTime()
		params = append(params, model.InsertOHLCVAsPerDayParams{
			SecID:  secID,
			Date:   pgtype.Date{Time: t, Valid: true},
			Open:   floatToNumeric(o.GetOpen()),
			High:   floatToNumeric(o.GetHigh()),
			Low:    floatToNumeric(o.GetLow()),
			Close:  floatToNumeric(o.GetClose()),
			Volume: int64(o.GetVolume()),
			Amount: uint64ToNumeric(o.GetAmount()),
		})
	}

	if _, err := s.db.InsertOHLCVAsPerDay(ctx, params); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) createOHLCVAsPerMin(ctx context.Context, secID pgtype.UUID, data []*pb.OHLCVA) (*emptypb.Empty, error) {
	minParams := make([]model.InsertOHLCVAsPerMinParams, 0, len(data))
	for _, o := range data {
		if o.GetTs() == nil {
			return nil, status.Error(codes.InvalidArgument, "timestamp is required")
		}
		t := o.GetTs().AsTime().Truncate(time.Minute)
		minParams = append(minParams, model.InsertOHLCVAsPerMinParams{
			SecID:  secID,
			Ts:     pgtype.Timestamp{Time: t, Valid: true},
			Open:   floatToNumeric(o.GetOpen()),
			High:   floatToNumeric(o.GetHigh()),
			Low:    floatToNumeric(o.GetLow()),
			Close:  floatToNumeric(o.GetClose()),
			Volume: int64(o.GetVolume()),
			Amount: uint64ToNumeric(o.GetAmount()),
		})
	}
	if _, err := s.db.InsertOHLCVAsPerMin(ctx, minParams); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Aggregate into 30-minute buckets
	type bucket struct {
		ts     time.Time
		open   float64
		high   float64
		low    float64
		close  float64
		volume uint64
		amount uint64
	}

	bucketMap := make(map[time.Time]*bucket)
	var bucketOrder []time.Time

	for _, o := range data {
		t := o.GetTs().AsTime().Truncate(time.Minute)
		k := t.Truncate(30 * time.Minute)
		if b, ok := bucketMap[k]; !ok {
			bucketMap[k] = &bucket{
				ts:     k,
				open:   o.GetOpen(),
				high:   o.GetHigh(),
				low:    o.GetLow(),
				close:  o.GetClose(),
				volume: o.GetVolume(),
				amount: o.GetAmount(),
			}
			bucketOrder = append(bucketOrder, k)
		} else {
			if o.GetHigh() > b.high {
				b.high = o.GetHigh()
			}
			if o.GetLow() < b.low {
				b.low = o.GetLow()
			}
			b.close = o.GetClose()
			b.volume += o.GetVolume()
			b.amount += o.GetAmount()
		}
	}

	thirtyMinParams := make([]model.InsertOHLCVAsPer30MinParams, 0, len(bucketOrder))
	for _, k := range bucketOrder {
		b := bucketMap[k]
		thirtyMinParams = append(thirtyMinParams, model.InsertOHLCVAsPer30MinParams{
			SecID:  secID,
			Ts:     pgtype.Timestamp{Time: b.ts, Valid: true},
			Open:   floatToNumeric(b.open),
			High:   floatToNumeric(b.high),
			Low:    floatToNumeric(b.low),
			Close:  floatToNumeric(b.close),
			Volume: int64(b.volume),
			Amount: uint64ToNumeric(b.amount),
		})
	}
	if _, err := s.db.InsertOHLCVAsPer30Min(ctx, thirtyMinParams); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) GetOHLCVAs(req *pb.GetOHLCVAsRequest, stream grpc.ServerStreamingServer[pb.OHLCVA]) error {
	ctx := stream.Context()

	if req.GetExchange() == "" || req.GetSymbol() == "" {
		return status.Error(codes.InvalidArgument, "exchange and symbol are required")
	}
	if req.Interval == nil || req.From == nil || req.Before == nil {
		return status.Error(codes.InvalidArgument, "interval, from, and before are required")
	}

	intervalSecs := req.Interval.Seconds
	switch intervalSecs {
	case interval1m, interval5m, interval30m, interval1h, interval1d, interval1w, interval1M:
	default:
		return status.Error(codes.InvalidArgument, "invalid interval")
	}

	fromTime := req.From.AsTime()
	beforeTime := req.Before.AsTime()
	if !fromTime.Before(beforeTime) {
		return status.Error(codes.InvalidArgument, "from must be earlier than before")
	}

	secs, err := s.db.GetSecuritiesBySymbols(ctx, req.GetExchange(), req.GetSymbol())
	if err != nil || len(secs) == 0 {
		return status.Error(codes.NotFound, "security not found")
	}
	secID := secs[0].ID

	switch intervalSecs {
	case interval1m, interval5m:
		return s.streamOHLCVAsPerMin(ctx, stream, secID, fromTime, beforeTime, intervalSecs)
	case interval30m, interval1h:
		return s.streamOHLCVAsPer30Min(ctx, stream, secID, fromTime, beforeTime, intervalSecs)
	default:
		return s.streamOHLCVAsPerDay(ctx, stream, secID, fromTime, beforeTime, intervalSecs)
	}
}

func (s *Service) streamOHLCVAsPerMin(
	ctx context.Context,
	stream grpc.ServerStreamingServer[pb.OHLCVA],
	secID pgtype.UUID,
	from, before time.Time,
	intervalSecs int64,
) error {
	rows, err := s.db.GetOHLCVAsPerMin(ctx, secID,
		pgtype.Timestamp{Time: from, Valid: true},
		pgtype.Timestamp{Time: before, Valid: true},
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if intervalSecs == interval1m {
		for _, row := range rows {
			if err := stream.Send(ohlcvaProto(row.Ts.Time, row.Open, row.High, row.Low, row.Close, row.Volume, row.Amount)); err != nil {
				return err
			}
		}
		return nil
	}

	// 5m: aggregate 1m rows into 5-minute buckets
	return streamAggregated(stream, rows, func(r model.OHLCVAperMin) (time.Time, ohlcvaValues) {
		return r.Ts.Time.Truncate(5 * time.Minute), ohlcvaValues{
			open:   numericToFloat(r.Open),
			high:   numericToFloat(r.High),
			low:    numericToFloat(r.Low),
			close:  numericToFloat(r.Close),
			volume: uint64(r.Volume),
			amount: numericToUint64(r.Amount),
		}
	})
}

func (s *Service) streamOHLCVAsPer30Min(
	ctx context.Context,
	stream grpc.ServerStreamingServer[pb.OHLCVA],
	secID pgtype.UUID,
	from, before time.Time,
	intervalSecs int64,
) error {
	rows, err := s.db.GetOHLCVAsPer30Min(ctx, secID,
		pgtype.Timestamp{Time: from, Valid: true},
		pgtype.Timestamp{Time: before, Valid: true},
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if intervalSecs == interval30m {
		for _, row := range rows {
			if err := stream.Send(ohlcvaProto(row.Ts.Time, row.Open, row.High, row.Low, row.Close, row.Volume, row.Amount)); err != nil {
				return err
			}
		}
		return nil
	}

	// 1h: aggregate 30m rows into 1-hour buckets
	return streamAggregated(stream, rows, func(r model.OHLCVAper30Min) (time.Time, ohlcvaValues) {
		return r.Ts.Time.Truncate(time.Hour), ohlcvaValues{
			open:   numericToFloat(r.Open),
			high:   numericToFloat(r.High),
			low:    numericToFloat(r.Low),
			close:  numericToFloat(r.Close),
			volume: uint64(r.Volume),
			amount: numericToUint64(r.Amount),
		}
	})
}

func (s *Service) streamOHLCVAsPerDay(
	ctx context.Context,
	stream grpc.ServerStreamingServer[pb.OHLCVA],
	secID pgtype.UUID,
	from, before time.Time,
	intervalSecs int64,
) error {
	rows, err := s.db.GetOHLCVAsPerDay(ctx, secID,
		pgtype.Date{Time: from, Valid: true},
		pgtype.Date{Time: before, Valid: true},
	)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if intervalSecs == interval1d {
		for _, row := range rows {
			if err := stream.Send(ohlcvaProto(row.Date.Time, row.Open, row.High, row.Low, row.Close, row.Volume, row.Amount)); err != nil {
				return err
			}
		}
		return nil
	}

	bucketFn := weekBucket
	if intervalSecs == interval1M {
		bucketFn = monthBucket
	}

	return streamAggregated(stream, rows, func(r model.OHLCVAperDay) (time.Time, ohlcvaValues) {
		return bucketFn(r.Date.Time), ohlcvaValues{
			open:   numericToFloat(r.Open),
			high:   numericToFloat(r.High),
			low:    numericToFloat(r.Low),
			close:  numericToFloat(r.Close),
			volume: uint64(r.Volume),
			amount: numericToUint64(r.Amount),
		}
	})
}

// weekBucket returns the Monday of the week containing t (ISO week start).
func weekBucket(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday → 7
	}
	monday := t.AddDate(0, 0, -(weekday - 1))
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, monday.Location())
}

// monthBucket returns the first day of the month containing t.
func monthBucket(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// ohlcvaValues holds float64 OHLCVA values for aggregation.
type ohlcvaValues struct {
	open   float64
	high   float64
	low    float64
	close  float64
	volume uint64
	amount uint64
}

// ohlcvaProto builds a pb.OHLCVA message from DB row fields.
func ohlcvaProto(ts time.Time, open, high, low, close_ pgtype.Numeric, volume int64, amount pgtype.Numeric) *pb.OHLCVA {
	o := numericToFloat(open)
	h := numericToFloat(high)
	l := numericToFloat(low)
	c := numericToFloat(close_)
	v := uint64(volume)
	a := numericToUint64(amount)
	return &pb.OHLCVA{
		Ts:     timestamppb.New(ts),
		Open:   &o,
		High:   &h,
		Low:    &l,
		Close:  &c,
		Volume: &v,
		Amount: &a,
	}
}

// streamAggregated groups rows by bucket key, aggregates OHLCVA values, and streams results.
func streamAggregated[R any](
	stream grpc.ServerStreamingServer[pb.OHLCVA],
	rows []R,
	extract func(R) (time.Time, ohlcvaValues),
) error {
	type agg struct {
		ts     time.Time
		open   float64
		high   float64
		low    float64
		close  float64
		volume uint64
		amount uint64
	}

	buckets := make(map[time.Time]*agg)
	var order []time.Time

	for _, row := range rows {
		k, v := extract(row)
		if b, ok := buckets[k]; !ok {
			buckets[k] = &agg{
				ts:     k,
				open:   v.open,
				high:   v.high,
				low:    v.low,
				close:  v.close,
				volume: v.volume,
				amount: v.amount,
			}
			order = append(order, k)
		} else {
			if v.high > b.high {
				b.high = v.high
			}
			if v.low < b.low {
				b.low = v.low
			}
			b.close = v.close
			b.volume += v.volume
			b.amount += v.amount
		}
	}

	for _, k := range order {
		b := buckets[k]
		if err := stream.Send(&pb.OHLCVA{
			Ts:     timestamppb.New(b.ts),
			Open:   &b.open,
			High:   &b.high,
			Low:    &b.low,
			Close:  &b.close,
			Volume: &b.volume,
			Amount: &b.amount,
		}); err != nil {
			return err
		}
	}

	return nil
}
