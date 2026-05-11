package statistics

import (
	"context"

	"cloud.google.com/go/civil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

func (s *Service) CreateMarginTransactions(ctx context.Context, req *pb.CreateMarginTransactionsRequest) (*emptypb.Empty, error) {
	if req.GetExchange() == "" {
		return nil, status.Error(codes.InvalidArgument, "exchange is required")
	}

	date := req.GetDate()
	if date == nil {
		return nil, status.Error(codes.InvalidArgument, "date is required")
	}

	txs := req.GetMarginTransactions()
	if len(txs) == 0 {
		return nil, status.Error(codes.InvalidArgument, "margin_transactions is required")
	}

	for symbol, mt := range txs {
		if symbol == "" {
			return nil, status.Error(codes.InvalidArgument, "symbol is required for each entry")
		}
		if mt == nil {
			return nil, status.Errorf(codes.InvalidArgument, "margin_transaction is required for symbol %s", symbol)
		}
	}

	if err := s.model.CreateMarginTransactions(ctx, req.GetExchange(), civil.DateOf(date.AsTime()), txs); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetMarginTransactions(req *pb.GetMarginTransactionsRequest, stream grpc.ServerStreamingServer[pb.MarginTransaction]) error {
	if req.GetExchange() == "" || req.GetSymbol() == "" {
		return status.Error(codes.InvalidArgument, "exchange and symbol are required")
	}
	if req.GetFrom() == nil || req.GetBefore() == nil {
		return status.Error(codes.InvalidArgument, "from and before are required")
	}

	fromTime := req.GetFrom().AsTime()
	beforeTime := req.GetBefore().AsTime()
	if !fromTime.Before(beforeTime) {
		return status.Error(codes.InvalidArgument, "from must be earlier than before")
	}

	txs, err := s.model.GetMarginTransactions(stream.Context(), req.GetExchange(), req.GetSymbol(), fromTime, beforeTime)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		if err := stream.Send(tx); err != nil {
			return err
		}
	}
	return nil
}
