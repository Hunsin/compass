package quote

import (
	"context"
	"errors"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

func (s *Service) CreateExchange(ctx context.Context, ex *pb.Exchange) (*emptypb.Empty, error) {
	if ex.GetAbbr() == "" || ex.GetName() == "" || ex.GetTimezone() == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}
	if _, err := time.LoadLocation(ex.GetTimezone()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid timezone")
	}

	abbr := strings.ToLower(ex.GetAbbr())
	normalized := &pb.Exchange{Abbr: &abbr, Name: ex.Name, Timezone: ex.Timezone}
	if err := s.model.CreateExchange(ctx, normalized); err != nil {
		if errors.Is(err, quoteLib.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "exchange already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetExchanges(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.Exchange]) error {
	exchanges, err := s.model.GetExchanges(stream.Context())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for _, ex := range exchanges {
		if err := stream.Send(ex); err != nil {
			return err
		}
	}
	return nil
}
