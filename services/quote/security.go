package quote

import (
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

func (s *Service) CreateSecurities(stream grpc.ClientStreamingServer[pb.Security, emptypb.Empty]) error {
	var securities []*pb.Security
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
		securities = append(securities, sec)
	}
	if len(securities) == 0 {
		return status.Error(codes.InvalidArgument, "at least one security is required")
	}

	if err := s.model.CreateSecurities(stream.Context(), securities); err != nil {
		switch {
		case errors.Is(err, quoteLib.ErrNotFound):
			return status.Error(codes.NotFound, "exchange not found")
		case errors.Is(err, quoteLib.ErrAlreadyExists):
			return status.Error(codes.AlreadyExists, "security already exists")
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}
	return stream.SendAndClose(&emptypb.Empty{})
}

func (s *Service) GetSecurities(ex *pb.Exchange, stream grpc.ServerStreamingServer[pb.Security]) error {
	abbr := ex.GetAbbr()
	if abbr == "" {
		return status.Error(codes.InvalidArgument, "abbr is required")
	}

	secs, err := s.model.GetSecurities(stream.Context(), abbr)
	if err != nil {
		if errors.Is(err, quoteLib.ErrNotFound) {
			return status.Errorf(codes.NotFound, "exchange %q not found", abbr)
		}
		return status.Error(codes.Internal, err.Error())
	}

	for _, sec := range secs {
		if err := stream.Send(sec); err != nil {
			return err
		}
	}
	return nil
}
