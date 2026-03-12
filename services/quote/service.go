package quote

import (
	"github.com/rs/zerolog"
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// Service implements the gRPC QuoteService.
type Service struct {
	pb.UnimplementedQuoteServiceServer
	model quoteLib.Model
	log   zerolog.Logger
}

// New creates a new Service using the given model and logger.
func New(m quoteLib.Model, log zerolog.Logger) *Service {
	return &Service{model: m, log: log}
}
