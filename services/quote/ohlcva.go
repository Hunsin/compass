package quote

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
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
	if intervalSecs != quoteLib.Interval1m && intervalSecs != quoteLib.Interval1d {
		return nil, status.Error(codes.InvalidArgument, "interval must be 1m or 1d")
	}

	for _, o := range req.Ohlcva {
		if o.GetTs() == nil {
			return nil, status.Error(codes.InvalidArgument, "timestamp is required")
		}
	}

	if err := s.model.CreateOHLCVAs(ctx, req.GetExchange(), req.GetSymbol(), intervalSecs, req.Ohlcva); err != nil {
		if errors.Is(err, quoteLib.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "security not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetOHLCVAs(req *pb.GetOHLCVAsRequest, stream grpc.ServerStreamingServer[pb.OHLCVA]) error {
	if req.GetExchange() == "" || req.GetSymbol() == "" {
		return status.Error(codes.InvalidArgument, "exchange and symbol are required")
	}
	if req.Interval == nil || req.From == nil || req.Before == nil {
		return status.Error(codes.InvalidArgument, "interval, from, and before are required")
	}

	intervalSecs := req.Interval.Seconds
	switch intervalSecs {
	case quoteLib.Interval1m, quoteLib.Interval5m, quoteLib.Interval30m,
		quoteLib.Interval1h, quoteLib.Interval1d, quoteLib.Interval1w, quoteLib.Interval1M:
	default:
		return status.Error(codes.InvalidArgument, "invalid interval")
	}

	fromTime := req.From.AsTime()
	beforeTime := req.Before.AsTime()
	if !fromTime.Before(beforeTime) {
		return status.Error(codes.InvalidArgument, "from must be earlier than before")
	}

	ohlcvas, err := s.model.GetOHLCVAs(stream.Context(), req.GetExchange(), req.GetSymbol(), intervalSecs, fromTime, beforeTime)
	if err != nil {
		if errors.Is(err, quoteLib.ErrNotFound) {
			return status.Error(codes.NotFound, "security not found")
		}
		return status.Error(codes.Internal, err.Error())
	}

	for _, o := range ohlcvas {
		if err := stream.Send(o); err != nil {
			return err
		}
	}
	return nil
}
