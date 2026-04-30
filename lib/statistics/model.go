// Package statistics provides domain operations for the Statistics service.
package statistics

import (
	"context"
	"time"

	pb "github.com/Hunsin/compass/protocols/gen/go/statistics/v1"
)

// Model defines the domain operations for the Statistics service.
type Model interface {
	// CreateMarginTransactions stores daily margin trading data for a security.
	// Returns an oops.NotFound error if the security does not exist.
	CreateMarginTransactions(ctx context.Context, exchange, symbol string, txs []*pb.MarginTransaction) error
	// GetMarginTransactions retrieves daily margin transactions for a security within [from, before).
	// Returns an oops.NotFound error if the security does not exist.
	GetMarginTransactions(ctx context.Context, exchange, symbol string, from, before time.Time) ([]*pb.MarginTransaction, error)
	// Health checks connectivity to the underlying database.
	Health(ctx context.Context) error
}
