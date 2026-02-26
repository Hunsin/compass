package quote

import (
	quoteLib "github.com/Hunsin/compass/lib/quote"
	pb "github.com/Hunsin/compass/protocols/gen/go/quote"
)

// Service implements the gRPC QuoteService.
type Service struct {
	pb.UnimplementedQuoteServiceServer
	model quoteLib.Model
}

// New creates a new Service using the given model.
func New(m quoteLib.Model) *Service {
	return &Service{model: m}
}
