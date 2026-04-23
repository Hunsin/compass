package quote

import (
	"context"
	"time"

	pb "github.com/Hunsin/compass/protocols/gen/go/quote/v1"
)

// Interval constants define supported OHLCV time intervals in seconds.
const (
	Interval1m  int64 = 60
	Interval5m  int64 = 300
	Interval30m int64 = 1800
	Interval1h  int64 = 3600
	Interval1d  int64 = 86400
	Interval1w  int64 = 604800
	Interval1M  int64 = 2592000
)

// Model defines the domain operations for the Quote service.
type Model interface {
	CreateExchange(ctx context.Context, ex *pb.Exchange) error
	GetExchanges(ctx context.Context) ([]*pb.Exchange, error)
	CreateSecurities(ctx context.Context, securities []*pb.Security) error
	// GetSecurities returns all securities for the given exchange abbreviation.
	// Returns an oops.NotFound error if the exchange does not exist.
	GetSecurities(ctx context.Context, exchange string) ([]*pb.Security, error)
	// CreateOHLCVs stores OHLCV data. interval must be Interval1m or Interval1d.
	// Returns an oops.InvalidArgument error if the interval is not supported.
	CreateOHLCVs(ctx context.Context, exchange, symbol string, interval int64, ohlcvs []*pb.OHLCV) error
	// GetOHLCVs retrieves OHLCV data aggregated to the requested interval.
	// Returns an oops.InvalidArgument error if the interval is not supported.
	GetOHLCVs(ctx context.Context, exchange, symbol string, interval int64, from, before time.Time) ([]*pb.OHLCV, error)
	// Health checks connectivity to underlying dependencies (database and cache).
	Health(ctx context.Context) error
}
