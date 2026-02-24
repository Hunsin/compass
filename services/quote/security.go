package quote

import (
	"errors"
	"io"

	"github.com/jackc/pgx/v5/pgconn"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/Hunsin/compass/postgres/gen/model"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
)

func (s *Service) CreateSecurities(stream grpc.ClientStreamingServer[pb.Security, emptypb.Empty]) error {
	ctx := stream.Context()

	exchanges, err := s.db.GetExchanges(ctx)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	exchangeSet := make(map[string]bool, len(exchanges))
	for _, ex := range exchanges {
		exchangeSet[ex.Abbr] = true
	}

	var params []model.InsertSecuritiesParams
	for {
		sec, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if sec.GetExchange() == "" || sec.GetSymbol() == "" || sec.GetName() == "" {
			return status.Error(codes.InvalidArgument, "all fields are required")
		}

		if !exchangeSet[sec.GetExchange()] {
			return status.Errorf(codes.NotFound, "exchange %q not found", sec.GetExchange())
		}

		params = append(params, model.InsertSecuritiesParams{
			Exchange: sec.GetExchange(),
			Symbol:   sec.GetSymbol(),
			Name:     sec.GetName(),
		})
	}

	if _, err := s.db.InsertSecurities(ctx, params); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return status.Error(codes.AlreadyExists, "security already exists")
			case "23503":
				return status.Error(codes.NotFound, "exchange not found")
			}
		}
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(&emptypb.Empty{})
}

func (s *Service) GetSecurities(ex *pb.Exchange, stream grpc.ServerStreamingServer[pb.Security]) error {
	ctx := stream.Context()
	abbr := ex.GetAbbr()
	if abbr == "" {
		return status.Error(codes.InvalidArgument, "abbr is required")
	}

	secs, err := s.db.GetSecurities(ctx, abbr)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if len(secs) == 0 {
		if _, err := s.db.GetExchange(ctx, abbr); err != nil {
			return status.Errorf(codes.NotFound, "exchange %q not found", abbr)
		}
		return nil
	}

	for _, sec := range secs {
		exch := sec.Exchange
		sym := sec.Symbol
		name := sec.Name
		if err := stream.Send(&pb.Security{Exchange: &exch, Symbol: &sym, Name: &name}); err != nil {
			return err
		}
	}

	return nil
}
