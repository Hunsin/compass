package quote

import (
	"errors"

	"github.com/rs/zerolog"

	"github.com/Hunsin/compass/lib/oops"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// Service implements the gRPC QuoteService.
type Service struct {
	pb.UnimplementedQuoteServiceServer
	model quoteLib.Model
	log   zerolog.Logger
}

// fromError converts a domain error to a gRPC status error.
func (s *Service) fromError(err error) error {
	var e *oops.Error
	if errors.As(err, &e) {
		return e.GRPC()
	}
	return err
}

// New creates a new Service using the given model and logger.
func New(m quoteLib.Model, log zerolog.Logger) *Service {
	return &Service{model: m, log: log}
}
