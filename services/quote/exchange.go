package quote

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

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
		return nil, s.fromError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetExchanges(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.Exchange]) error {
	exchanges, err := s.model.GetExchanges(stream.Context())
	if err != nil {
		return s.fromError(err)
	}
	for _, ex := range exchanges {
		if err := stream.Send(ex); err != nil {
			return err
		}
	}
	return nil
}
