package statistics

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

func (s *Service) CreateMarginTransactions(ctx context.Context, req *pb.CreateMarginTransactionsRequest) (*emptypb.Empty, error) {
	if req.GetExchange() == "" || req.GetSymbol() == "" {
		return nil, status.Error(codes.InvalidArgument, "exchange and symbol are required")
	}
	if len(req.GetMarginTransactions()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "margin_transactions is required")
	}
	for _, tx := range req.GetMarginTransactions() {
		if tx.GetDate() == nil {
			return nil, status.Error(codes.InvalidArgument, "date is required for each transaction")
		}
	}

	if err := s.model.CreateMarginTransactions(ctx, req.GetExchange(), req.GetSymbol(), req.GetMarginTransactions()); err != nil {
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
