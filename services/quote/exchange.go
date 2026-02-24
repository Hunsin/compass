package quote

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
)

func (s *Service) CreateExchange(ctx context.Context, ex *pb.Exchange) (*emptypb.Empty, error) {
	if ex.GetAbbr() == "" || ex.GetName() == "" || ex.GetTimezone() == "" {
		return nil, status.Error(codes.InvalidArgument, "all fields are required")
	}

	if _, err := time.LoadLocation(ex.GetTimezone()); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid timezone")
	}

	abbr := strings.ToLower(ex.GetAbbr())
	if err := s.db.InsertExchange(ctx, abbr, ex.GetName(), ex.GetTimezone()); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, status.Error(codes.AlreadyExists, "exchange already exists")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) GetExchanges(_ *emptypb.Empty, stream grpc.ServerStreamingServer[pb.Exchange]) error {
	exchanges, err := s.db.GetExchanges(stream.Context())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	for _, ex := range exchanges {
		abbr := ex.Abbr
		name := ex.Name
		tz := ex.Timezone
		if err := stream.Send(&pb.Exchange{Abbr: &abbr, Name: &name, Timezone: &tz}); err != nil {
			return err
		}
	}

	return nil
}
