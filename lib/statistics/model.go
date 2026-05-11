// Package statistics provides domain operations for the Statistics service.
package statistics

import (
	"context"
	"time"

	"cloud.google.com/go/civil"

	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

// Model defines the domain operations for the Statistics service.
type Model interface {
	// CreateMarginTransactions stores daily margin trading data for the given
	// date across multiple securities of an exchange. The map key is the symbol.
	// Returns an oops.NotFound error if any security does not exist.
	CreateMarginTransactions(ctx context.Context, exchange string, date civil.Date, txs map[string]*pb.MarginTransaction) error
	// GetMarginTransactions retrieves daily margin transactions for a security within [from, before).
	// Returns an oops.NotFound error if the security does not exist.
	GetMarginTransactions(ctx context.Context, exchange, symbol string, from, before time.Time) ([]*pb.MarginTransaction, error)
	// Health checks connectivity to the underlying database.
	Health(ctx context.Context) error
}
